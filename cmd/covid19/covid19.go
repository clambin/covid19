package main

import (
	"context"
	"fmt"
	"github.com/clambin/covid19/configuration"
	"github.com/clambin/covid19/covidcache"
	"github.com/clambin/covid19/coviddb"
	"github.com/clambin/covid19/covidhandler"
	"github.com/clambin/covid19/covidprobe"
	"github.com/clambin/covid19/db"
	popdb "github.com/clambin/covid19/population/db"
	"github.com/clambin/covid19/population/probe"
	"github.com/clambin/covid19/version"
	"github.com/clambin/grafana-json"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"github.com/xonvanetta/shutdown/pkg/shutdown"
	"gopkg.in/alecthomas/kingpin.v2"
	"net/http"
	_ "net/http/pprof"
	"os"
	"path/filepath"
	"time"
)

func main() {
	var (
		cfg            *configuration.Configuration
		debug          bool
		configFileName string
	)

	log.WithField("version", version.BuildVersion).Info("covid19 monitor starting")
	a := kingpin.New(filepath.Base(os.Args[0]), "covid19 monitor")

	a.Version(version.BuildVersion)
	a.HelpFlag.Short('h')
	a.VersionFlag.Short('v')
	a.Flag("debug", "Log debug messages").BoolVar(&debug)
	a.Flag("config", "Configuration file").Required().StringVar(&configFileName)

	_, err := a.Parse(os.Args[1:])
	if err != nil {
		a.Usage(os.Args[1:])
		os.Exit(2)
	}

	if cfg, err = configuration.LoadConfigurationFile(configFileName); err != nil {
		log.WithField("err", err).Fatal("Failed to read config file")
	}

	if debug || cfg.Debug {
		log.SetLevel(log.DebugLevel)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	covidDB, populationDB := openDatabases(cfg)
	cache := covidcache.New(covidDB)
	go cache.Run(ctx)

	populationProbe := probe.Create(cfg.Monitor.RapidAPIKey.Value, populationDB, covidDB)
	populationTicker := time.NewTicker(24 * time.Hour)

	covidProbe := covidprobe.NewProbe(&cfg.Monitor, covidDB, cache)
	prometheus.MustRegister(covidProbe)
	covidTicker := time.NewTicker(cfg.Monitor.Interval)

	server := startAPIServer(cache, cfg)

	for running := true; running; {
		select {
		case <-ctx.Done():
			running = false
		case <-shutdown.Chan():
			running = false
		case <-populationTicker.C:
			populationProbe.Update(ctx)
		case <-covidTicker.C:
			_ = covidProbe.Update(ctx)
		}
	}
	covidTicker.Stop()
	populationTicker.Stop()
	_ = server.Shutdown(context.Background())
}

func startAPIServer(cache *covidcache.Cache, cfg *configuration.Configuration) (httpServer *http.Server) {
	server := grafana_json.Server{
		Handler: covidhandler.Create(cache),
	}
	r := server.GetRouter()
	r.PathPrefix("/debug/pprof/").Handler(http.DefaultServeMux)

	httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: r,
	}

	go func() {
		err := httpServer.ListenAndServe()
		if err != http.ErrServerClosed {
			log.WithError(err).Fatal("unable to start grafana SimpleJson server")
		}
	}()

	return
}

func openDatabases(cfg *configuration.Configuration) (covidDB coviddb.DB, populationDB popdb.DB) {
	DB, err := db.New(
		cfg.Postgres.Host,
		cfg.Postgres.Port,
		cfg.Postgres.Database,
		cfg.Postgres.User,
		cfg.Postgres.Password,
	)

	if err != nil {
		log.WithError(err).Fatalf("unable to access covid DB '%s'", cfg.Postgres.Database)
	}

	covidDB, err = coviddb.New(DB)

	if err != nil {
		log.WithError(err).Fatalf("unable to access covid DB '%s'", cfg.Postgres.Database)
	}

	populationDB, err = popdb.New(DB)

	if err != nil {
		log.WithError(err).Fatalf("unable to access population DB '%s'", cfg.Postgres.Database)
	}
	return
}
