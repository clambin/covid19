package main

import (
	"os"
	"path/filepath"
	// "runtime/pprof"

	log "github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"

	"covid19/internal/coviddb"
	"covid19/internal/covidhandler"
	"covid19/internal/version"
	"covid19/pkg/grafana/apiserver"
)

func main() {
	cfg := struct {
		port             int
		debug            bool
		postgresHost     string
		postgresPort     int
		postgresDatabase string
		postgresUser     string
		postgresPassword string
	}{}

	a := kingpin.New(filepath.Base(os.Args[0]), "covid19 grafana API server")

	a.Version(version.BuildVersion)
	a.HelpFlag.Short('h')
	a.VersionFlag.Short('v')
	a.Flag("port", "API listener port").Default("5000").IntVar(&cfg.port)
	a.Flag("debug", "Log debug messages").BoolVar(&cfg.debug)
	a.Flag("postgres-host", "Postgres DB Host").Default("postgres").StringVar(&cfg.postgresHost)
	a.Flag("postgres-port", "Postgres DB Port").Default("5432").IntVar(&cfg.postgresPort)
	a.Flag("postgres-database", "Postgres DB Name").Default("covid19").StringVar(&cfg.postgresDatabase)
	a.Flag("postgres-user", "Postgres DB User").Default("covid").StringVar(&cfg.postgresUser)
	a.Flag("postgres-password", "Postgres DB Password").StringVar(&cfg.postgresPassword)

	_, err := a.Parse(os.Args[1:])
	if err != nil {
		a.Usage(os.Args[1:])
		os.Exit(2)
	}

	if cfg.debug {
		log.SetLevel(log.DebugLevel)
	}

	log.Info("covid19api v" + version.BuildVersion)

	covidDB := coviddb.NewPostgresDB(cfg.postgresHost, cfg.postgresPort, cfg.postgresDatabase, cfg.postgresUser, cfg.postgresPassword)
	handler, _ := covidhandler.Create(covidDB)
	server := apiserver.Create(handler, cfg.port)
	server.Run()
}
