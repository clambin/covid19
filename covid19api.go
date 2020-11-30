package main

import (
	// "fmt"
	"path/filepath"
	"os"

	kingpin "gopkg.in/alecthomas/kingpin.v2"

	"covid19api/coviddb"
	"covid19api/apiserver"
)

func main() {
	cfg := struct {
		port              int
		debug             bool
		postgres_host     string
		postgres_port     int
		postgres_database string
		postgres_user     string
		postgres_password string
	}{}

	a := kingpin.New(filepath.Base(os.Args[0]), "covid19 grafana API server")

	a.HelpFlag.Short('h')

	a.Flag("port", "API listener port").Default("5000").IntVar(&cfg.port)
	a.Flag("postgres-host", "Postgres DB Host").Default("postgres").StringVar(&cfg.postgres_host)
	a.Flag("postgres-port", "Postgres DB Port").Default("5432").IntVar(&cfg.postgres_port)
	a.Flag("postgres-database", "Postgres DB Name").Default("covid19").StringVar(&cfg.postgres_database)
	a.Flag("postgres-user", "Postgres DB User").Default("covid").StringVar(&cfg.postgres_user)
	a.Flag("postgres-password", "Postgres DB Password").StringVar(&cfg.postgres_password)

	_, err := a.Parse(os.Args[1:])
	if err != nil {
		a.Usage(os.Args[1:])
		os.Exit(2)
	}

	db := coviddb.Create(cfg.postgres_host, cfg.postgres_port, cfg.postgres_database, cfg.postgres_user, cfg.postgres_password)

	server := apiserver.Server(apiserver.Handler(db))

	server.Run()
}


