package main

import (
	"os"
	"path/filepath"
	"runtime/pprof"
	"time"

	log "github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"

	covidapi "covid19/internal/covid/apiclient"
	covidprobe "covid19/internal/covid/probe"
	"covid19/internal/coviddb"
	popapi "covid19/internal/population/apiclient"
	popdb "covid19/internal/population/db"
	popprobe "covid19/internal/population/probe"
	"covid19/internal/pushgateway"
	"covid19/internal/version"
)

func main() {
	cfg := struct {
		debug            bool
		once             bool
		interval         time.Duration
		postgresHost     string
		postgresPort     int
		postgresDatabase string
		postgresUser     string
		postgresPassword string
		apiKey           string
		pushGateway      string
		profileName      string
	}{}

	a := kingpin.New(filepath.Base(os.Args[0]), "covid19 monitor")

	a.Version(version.BuildVersion)
	a.HelpFlag.Short('h')
	a.VersionFlag.Short('v')
	a.Flag("debug", "Log debug messages").BoolVar(&cfg.debug)
	a.Flag("once", "Run once and then exit").BoolVar(&cfg.once)
	a.Flag("interval", "Time between measurements").Default("20m").DurationVar(&cfg.interval)
	a.Flag("postgres-host", "Postgres DB Host").Default("postgres").StringVar(&cfg.postgresHost)
	a.Flag("postgres-port", "Postgres DB Port").Default("5432").IntVar(&cfg.postgresPort)
	a.Flag("postgres-database", "Postgres DB Name").Default("covid19").StringVar(&cfg.postgresDatabase)
	a.Flag("postgres-user", "Postgres DB User").Default("covid").StringVar(&cfg.postgresUser)
	a.Flag("postgres-password", "Postgres DB Password").StringVar(&cfg.postgresPassword)
	a.Flag("api-key", "API Key for RapidAPI Covid19 API").StringVar(&cfg.apiKey)
	a.Flag("pushgateway", "URL of Prometheus pushgateway").StringVar(&cfg.pushGateway)
	a.Flag("profile", "Filename for go profiler").StringVar(&cfg.profileName)

	if _, err := a.Parse(os.Args[1:]); err != nil {
		a.Usage(os.Args[1:])
		os.Exit(1)
	}

	if cfg.debug {
		log.SetLevel(log.DebugLevel)
	}

	log.Info("covid19mon v" + version.BuildVersion)

	if cfg.profileName != "" {
		f, ferr := os.Create(cfg.profileName)
		if ferr != nil {
			log.Fatal(ferr)
		}
		_ = pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	covidProbe := covidprobe.NewProbe(
		covidapi.New(cfg.apiKey),
		coviddb.NewPostgresDB(cfg.postgresHost, cfg.postgresPort, cfg.postgresDatabase, cfg.postgresUser, cfg.postgresPassword),
		pushgateway.NewPushGateway(cfg.pushGateway))

	populationProbe := popprobe.Create(
		popapi.New(cfg.apiKey),
		popdb.NewPostgresDB(
			cfg.postgresHost, cfg.postgresPort, cfg.postgresDatabase, cfg.postgresUser, cfg.postgresPassword))

	if cfg.once {
		if err := covidProbe.Run(); err != nil {
			log.Warningf("covid probe error: %s", err)
		}
		if err := populationProbe.Run(); err != nil {
			log.Warningf("covid probe error: %s", err)
		}
	} else {
		go func() {
			for {
				if err := covidProbe.Run(); err != nil {
					log.Warningf("covid probe error: %s", err)
				}
				time.Sleep(cfg.interval)
			}
		}()
		go func() {
			for {
				if err := populationProbe.Run(); err != nil {
					log.Warningf("covid probe error: %s", err)
				}
				time.Sleep(cfg.interval)
			}
		}()

		for {
			time.Sleep(cfg.interval)
		}
	}
}
