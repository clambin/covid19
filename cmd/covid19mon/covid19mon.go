package covid19mon

import (
	"os"
	"time"
	"path/filepath"

	kingpin "gopkg.in/alecthomas/kingpin.v2"
	log     "github.com/sirupsen/logrus"

	"covid19api/internal/coviddb"
	"covid19api/internal/scheduler"
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
	}{}

	a := kingpin.New(filepath.Base(os.Args[0]), "covid19 grafana API server")

	a.HelpFlag.Short('h')
	a.Flag("debug", "Log debug messages").BoolVar(&cfg.debug)
	a.Flag("once", "Run once and then exit").BoolVar(&cfg.once)
	a.Flag("interval", "Time between measurements").Default("1200").IntVar(&cfg.interval)
	a.Flag("postgres-host", "Postgres DB Host").Default("postgres").StringVar(&cfg.postgresHost)
	a.Flag("postgres-port", "Postgres DB Port").Default("5432").IntVar(&cfg.postgresPort)
	a.Flag("postgres-database", "Postgres DB Name").Default("covid19").StringVar(&cfg.postgresDatabase)
	a.Flag("postgres-user", "Postgres DB User").Default("covid").StringVar(&cfg.postgresUser)
	a.Flag("postgres-password", "Postgres DB Password").StringVar(&cfg.postgresPassword)
	a.Flag("api-key", "API Key for RapidAPI Covid19 API").StringVar(&cfg.apiKey)
	a.Flag("push-gateway", "URL of Prometheus pushgateway").StringVar(&cfg.pushGateway)

	if _, err := a.Parse(os.Args[1:]); err != nil {
		a.Usage(os.Args[1:])
		os.Exit(1)
	}

	if cfg.debug {
		log.SetLevel(log.DebugLevel)
	}

	db := coviddb.Create(cfg.postgresHost, cfg.postgresPort, cfg.postgresDatabase, cfg.postgresUser, cfg.postgresPassword)

	probe := covidprobe.Create(cfg.apiKey, cfg.pushGateway)
	scheduler := scheduler.NewScheduler()
	scheduler.Register(probe, time.Duration(cfg.interval) * time.Second)
	scheduler.Run(cfg.once)
}
