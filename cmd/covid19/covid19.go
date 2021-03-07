package main

import (
	"covid19/internal/configuration"
	"covid19/internal/covidcache"
	"covid19/internal/coviddb"
	"covid19/internal/covidhandler"
	"covid19/internal/covidprobe"
	popdb "covid19/internal/population/db"
	popprobe "covid19/internal/population/probe"
	"covid19/internal/version"
	"covid19/pkg/grafana/apiserver"
	"fmt"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
	"net/http"
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

	log.WithField("version", version.BuildVersion).Info("covid19 monitor starting")

	var cache *covidcache.Cache

	if cfg.Monitor.Enabled {
		cache = startMonitor(cfg, cfg.Grafana.Enabled)
	}

	if cfg.Grafana.Enabled {
		runGrafanaServer(cfg, cache)
	} else {
		// Grafana Server won't start prometheus server, so we start one manually
		listenAddress := fmt.Sprintf(":%d", cfg.Port)
		http.Handle("/metrics", promhttp.Handler())
		_ = http.ListenAndServe(listenAddress, nil)
	}
}

func startMonitor(cfg *configuration.Configuration, createCache bool) (cache *covidcache.Cache) {
	covidDB := coviddb.NewPostgresDB(
		cfg.Postgres.Host,
		cfg.Postgres.Port,
		cfg.Postgres.Database,
		cfg.Postgres.User,
		cfg.Postgres.Password,
	)

	popDB := popdb.NewPostgresDB(
		cfg.Postgres.Host,
		cfg.Postgres.Port,
		cfg.Postgres.Database,
		cfg.Postgres.User,
		cfg.Postgres.Password,
	)

	if createCache {
		cache = covidcache.New(covidDB)
		go cache.Run()
	}
	covidProbe := covidprobe.NewProbe(&cfg.Monitor, covidDB, cache)
	populationProbe := popprobe.Create(cfg.Monitor.RapidAPIKey.Value, popDB)

	// TODO: only update population once a day?
	go func() {
		var err error

		for {
			if err = covidProbe.Run(); err != nil {
				log.WithField("err", err).Warning("covidProbe failed")
			}
			if err = populationProbe.Run(); err != nil {
				log.WithField("err", err).Warning("populationProbe failed")
			}

			time.Sleep(cfg.Monitor.Interval)
		}
	}()

	return
}

func runGrafanaServer(cfg *configuration.Configuration, cache *covidcache.Cache) {
	handler, _ := covidhandler.Create(cache)
	server := apiserver.Create(handler, cfg.Port)
	_ = server.Run()
}
