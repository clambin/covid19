package main

import (
	covidapi "covid19/internal/covid/apiclient"
	covidprobe "covid19/internal/covid/probe"
	"covid19/internal/coviddb"
	"covid19/internal/monitor"
	popapi "covid19/internal/population/apiclient"
	popdb "covid19/internal/population/db"
	popprobe "covid19/internal/population/probe"
	"covid19/internal/reporters"
	log "github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
	"os"
	"path/filepath"
	"runtime/pprof"
	"strings"
	"time"

	"covid19/internal/version"
)

func main() {
	reportUpdatesCountries := ""

	cfg := monitor.Configuration{}

	a := kingpin.New(filepath.Base(os.Args[0]), "covid19 monitor")

	a.Version(version.BuildVersion)
	a.HelpFlag.Short('h')
	a.VersionFlag.Short('v')
	a.Flag("debug", "Log debug messages").BoolVar(&cfg.Debug)
	a.Flag("once", "Run once and then exit").BoolVar(&cfg.Once)
	a.Flag("interval", "Time between measurements").Default("20m").DurationVar(&cfg.Interval)
	a.Flag("postgres-host", "Postgres DB Host").Default("postgres").StringVar(&cfg.PostgresHost)
	a.Flag("postgres-port", "Postgres DB Port").Default("5432").IntVar(&cfg.PostgresPort)
	a.Flag("postgres-database", "Postgres DB Name").Default("covid19").StringVar(&cfg.PostgresDatabase)
	a.Flag("postgres-user", "Postgres DB User").Default("covid").StringVar(&cfg.PostgresUser)
	a.Flag("postgres-password", "Postgres DB Password").StringVar(&cfg.PostgresPassword)
	a.Flag("api-key", "API Key for RapidAPI Covid19 API").StringVar(&cfg.APIKey)
	a.Flag("pushgateway", "URL of Prometheus pushgateway (obsolete").StringVar(&cfg.Reports.Countries.PushGateway)
	a.Flag("report.countries.url", "URL of Prometheus pushgateway to report daily update for all countries").StringVar(&cfg.Reports.Countries.PushGateway)
	a.Flag("report.updates.token", "pushover API token to report selected country updates").StringVar(&cfg.Reports.Updates.Token)
	a.Flag("report.updates.user", "pushover user token to report selected country updates").StringVar(&cfg.Reports.Updates.User)
	a.Flag("report.updates.countries", "comma-separated list of countries to report").StringVar(&reportUpdatesCountries)
	a.Flag("profile", "Filename for go profiler").StringVar(&cfg.ProfileName)

	if _, err := a.Parse(os.Args[1:]); err != nil {
		a.Usage(os.Args[1:])
		os.Exit(1)
	}

	if reportUpdatesCountries != "" {
		cfg.Reports.Updates.Countries = strings.Split(reportUpdatesCountries, ",")
	}

	log.Info("covid19mon v" + version.BuildVersion)

	if cfg.ProfileName != "" {
		cfg.Once = true
		f, err := os.Create(cfg.ProfileName)
		if err != nil {
			log.Fatal(err)
		}
		_ = pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	covidDB := coviddb.NewPostgresDB(
		cfg.PostgresHost,
		cfg.PostgresPort,
		cfg.PostgresDatabase,
		cfg.PostgresUser,
		cfg.PostgresPassword,
	)

	popDB := popdb.NewPostgresDB(
		cfg.PostgresHost,
		cfg.PostgresPort,
		cfg.PostgresDatabase,
		cfg.PostgresUser,
		cfg.PostgresPassword,
	)

	rep := reporters.Create()

	if cfg.Reports.Countries.PushGateway != "" {
		rep.Add(reporters.NewCountriesReporter(cfg.Reports.Countries.PushGateway))
	}

	if cfg.Reports.Updates.Token != "" && cfg.Reports.Updates.User != "" {
		rep.Add(reporters.NewUpdatesReporter(
			cfg.Reports.Updates.Token,
			cfg.Reports.Updates.User,
			cfg.Reports.Updates.Countries,
			covidDB,
		),
		)
	}

	covidProbe := covidprobe.NewProbe(covidapi.New(cfg.APIKey), covidDB, rep)
	populationProbe := popprobe.Create(popapi.New(cfg.APIKey), popDB)

	for {
		if ok := monitor.Run(&cfg, covidProbe, populationProbe); !ok || cfg.Once {
			break
		}

		time.Sleep(cfg.Interval)
	}
}
