package main

import (
	"covid19/internal/coviddb"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"runtime/pprof"

	log "github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"

	"covid19/internal/covid"
	"covid19/internal/population"
	"covid19/internal/version"
	"covid19/pkg/scheduler"
)

func main() {
	cfg := struct {
		debug            bool
		once             bool
		interval         int
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
	a.Flag("interval", "Time between measurements").Default("1200").IntVar(&cfg.interval)
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

	newScheduler := scheduler.NewScheduler()

	// Add the covid probe
	covidProbe := covid.NewProbe(
		covid.NewAPIClient(&http.Client{}, cfg.apiKey),
		coviddb.NewPostgresDB(cfg.postgresHost, cfg.postgresPort, cfg.postgresDatabase, cfg.postgresUser, cfg.postgresPassword),
		covid.NewPushGateway(cfg.pushGateway))
	newScheduler.Register(covidProbe, time.Duration(cfg.interval)*time.Second)

	// Add the population probe
	populationProbe := population.Create(
		population.NewAPIClient(&http.Client{}, cfg.apiKey),
		population.NewPostgresDB(
			cfg.postgresHost, cfg.postgresPort, cfg.postgresDatabase, cfg.postgresUser, cfg.postgresPassword))
	newScheduler.Register(populationProbe, time.Duration(cfg.interval)*time.Second)

	// Go time
	newScheduler.Run(cfg.once)
}
