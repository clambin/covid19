package main

import (
	"path/filepath"
	"os"
	// "runtime/pprof"

	kingpin "gopkg.in/alecthomas/kingpin.v2"
	log     "github.com/sirupsen/logrus"

	"covid19/internal/apiserver"
	"covid19/internal/covid"
)

func main() {
	// f, ferr := os.Create("covid19api.prof")
	// if ferr != nil {
	// 	log.Fatal(ferr)
	// }
	// pprof.StartCPUProfile(f)
	// defer pprof.StopCPUProfile()

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

	a.HelpFlag.Short('h')

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

	db := covid.NewPGCovidDB(cfg.postgresHost, cfg.postgresPort, cfg.postgresDatabase, cfg.postgresUser, cfg.postgresPassword)
	handler := apiserver.CreateCovidAPIHandler(db)
	server := apiserver.CreateGrafanaAPIServer(handler, cfg.port)
	server.Run()
}
