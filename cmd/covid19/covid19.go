package main

import (
	"github.com/clambin/covid19/configuration"
	"github.com/clambin/covid19/covidcache"
	"github.com/clambin/covid19/coviddb"
	"github.com/clambin/covid19/covidhandler"
	"github.com/clambin/covid19/covidprobe"
	"github.com/clambin/covid19/population/db"
	"github.com/clambin/covid19/population/probe"
	"github.com/clambin/covid19/version"
	"github.com/clambin/grafana-json"
	log "github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
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

	var cache *covidcache.Cache
	if cfg.Monitor.Enabled {
		cache = startMonitor(cfg)
	}

	runGrafanaServer(cfg, cache)
}

func startMonitor(cfg *configuration.Configuration) (cache *covidcache.Cache) {
	covidDB := coviddb.NewPostgresDB(
		cfg.Postgres.Host,
		cfg.Postgres.Port,
		cfg.Postgres.Database,
		cfg.Postgres.User,
		cfg.Postgres.Password,
	)

	popDB := db.NewPostgresDB(
		cfg.Postgres.Host,
		cfg.Postgres.Port,
		cfg.Postgres.Database,
		cfg.Postgres.User,
		cfg.Postgres.Password,
	)

	cache = covidcache.New(covidDB)
	go cache.Run()

	populationProbe := probe.Create(cfg.Monitor.RapidAPIKey.Value, popDB, covidDB)
	go func() {
		err := populationProbe.Run()
		if err != nil {
			log.WithError(err).Warning("failed to get latest population figures")
		}
	}()

	covidProbe := covidprobe.NewProbe(&cfg.Monitor, covidDB, cache)

	go func() {
		for {
			err := covidProbe.Run()
			if err != nil {
				log.WithField("err", err).Warning("covidProbe failed")
			}
			time.Sleep(cfg.Monitor.Interval)
		}
	}()

	return
}

func runGrafanaServer(cfg *configuration.Configuration, cache *covidcache.Cache) {
	handler, _ := covidhandler.Create(cache)
	server := grafana_json.Create(handler, cfg.Port)
	_ = server.Run()
}
