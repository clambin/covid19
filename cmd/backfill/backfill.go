package main

import (
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"

	"covid19/internal/backfill"
	"covid19/internal/coviddb"
	"covid19/internal/version"
)

func main() {
	cfg := struct {
		debug            bool
		postgresHost     string
		postgresPort     int
		postgresDatabase string
		postgresUser     string
		postgresPassword string
	}{}

	a := kingpin.New(filepath.Base(os.Args[0]), "covid19 backfill tool")

	a.Version(version.BuildVersion)
	a.HelpFlag.Short('h')
	a.VersionFlag.Short('v')
	a.Flag("debug", "Log debug messages").BoolVar(&cfg.debug)
	a.Flag("postgres-host", "Postgres DB Host").Default("postgres").StringVar(&cfg.postgresHost)
	a.Flag("postgres-port", "Postgres DB Port").Default("5432").IntVar(&cfg.postgresPort)
	a.Flag("postgres-database", "Postgres DB Name").Default("covid19").StringVar(&cfg.postgresDatabase)
	a.Flag("postgres-user", "Postgres DB User").Default("covid").StringVar(&cfg.postgresUser)
	a.Flag("postgres-password", "Postgres DB Password").StringVar(&cfg.postgresPassword)

	if _, err := a.Parse(os.Args[1:]); err != nil {
		a.Usage(os.Args[1:])
		os.Exit(1)
	}

	if cfg.debug {
		log.SetLevel(log.DebugLevel)
	}

	log.WithField("version", version.BuildVersion).Info("backfill")

	app := backfill.Create(coviddb.NewPostgresDB(
		cfg.postgresHost,
		cfg.postgresPort,
		cfg.postgresDatabase,
		cfg.postgresUser,
		cfg.postgresPassword))

	err := app.Run()

	if err != nil {
		log.Error(err)
		os.Exit(1)
	}
}
