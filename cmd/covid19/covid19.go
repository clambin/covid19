package main

import (
	"context"
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
	log "github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
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

	var cache *covidcache.Cache
	if cfg.Monitor.Enabled {
		cache = startMonitor(ctx, cfg)
	}

	handler, _ := covidhandler.Create(cache)
	server := grafana_json.Create(handler)
	go func() {
		if err = server.Run(cfg.Port); err != nil {
			log.WithError(err).Fatal("unable to start grafana SimpleJson server")
		}
	}()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	<-sigs
}

func startMonitor(ctx context.Context, cfg *configuration.Configuration) (cache *covidcache.Cache) {
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

	var covidDB *coviddb.PostgresDB
	covidDB, err = coviddb.New(DB)

	if err != nil {
		log.WithError(err).Fatalf("unable to access covid DB '%s'", cfg.Postgres.Database)
	}

	var popDB *popdb.PostgresDB
	popDB, err = popdb.New(DB)

	if err != nil {
		log.WithError(err).Fatalf("unable to access population DB '%s'", cfg.Postgres.Database)
	}

	cache = covidcache.New(covidDB)
	go cache.Run(ctx)

	populationProbe := probe.Create(cfg.Monitor.RapidAPIKey.Value, popDB, covidDB)
	go func() {
		err2 := populationProbe.Run(ctx, 24*time.Hour)

		if err2 != nil {
			log.WithError(err2).Fatal("unable to get population data")
		}
	}()

	var covidProbe *covidprobe.Probe
	covidProbe, err = covidprobe.NewProbe(&cfg.Monitor, covidDB, cache)

	if err != nil {
		log.WithError(err).Fatal("failed to start covid probe")
	}
	go func() {
		err2 := covidProbe.Run(ctx, cfg.Monitor.Interval)

		if err2 != nil {
			log.WithError(err2).Fatal("unable to get covid data")
		}
	}()

	return
}
