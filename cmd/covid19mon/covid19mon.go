package main

import (
	"os"
	"path/filepath"
	"runtime/pprof"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"

	"covid19/internal/coviddb"
	"covid19/internal/covidprobe"
	"covid19/internal/monitor"
	popdb "covid19/internal/population/db"
	popprobe "covid19/internal/population/probe"
	"covid19/internal/reporters"
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
	a.Flag("pushgateway", "URL of Prometheus pushgateway to report daily update for all countries").StringVar(&cfg.PushGateway)
	a.Flag("profile", "Filename for go profiler").StringVar(&cfg.ProfileName)
	a.Flag("postgres.host", "Postgres DB Host").Default("postgres").StringVar(&cfg.Postgres.Host)
	a.Flag("postgres.port", "Postgres DB Port").Default("5432").IntVar(&cfg.Postgres.Port)
	a.Flag("postgres.database", "Postgres DB Name").Default("covid19").StringVar(&cfg.Postgres.Database)
	a.Flag("postgres.user", "Postgres DB User").Default("covid").StringVar(&cfg.Postgres.User)
	a.Flag("postgres.password", "Postgres DB Password").StringVar(&cfg.Postgres.Password)
	a.Flag("rapidapi.key", "API Key for RapidAPI Covid19 API").StringVar(&cfg.RapidAPI.Key)
	a.Flag("report.updates.countries", "comma-separated list of countries to report").StringVar(&reportUpdatesCountries)
	a.Flag("report.updates.pushover.token", "pushover API token to report selected country updates").StringVar(&cfg.Reports.Updates.Pushover.Token)
	a.Flag("report.updates.pushover.user", "pushover user token to report selected country updates").StringVar(&cfg.Reports.Updates.Pushover.User)
	a.Flag("report.updates.slack.url", "slack webhook URL to report selected country updates").StringVar(&cfg.Reports.Updates.Slack.URL)
	a.Flag("report.updates.slack.channel", "slack channel to report selected country updates").Default("#covid").StringVar(&cfg.Reports.Updates.Slack.Channel)

	if _, err := a.Parse(os.Args[1:]); err != nil {
		a.Usage(os.Args[1:])
		os.Exit(1)
	}

	if reportUpdatesCountries != "" {
		cfg.Reports.Countries = strings.Split(reportUpdatesCountries, ",")
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
		cfg.Postgres.Host,
		cfg.Postgres.Port,
		cfg.Postgres.Database,
		cfg.Postgres.User,
		cfg.Postgres.Password,
	)

	popDB := popdb.NewPostgresDB(
		cfg.Postgres.Host,
		cfg.Postgres.Port,
		cfg.Postgres.Database,
		cfg.Postgres.User,
		cfg.Postgres.Password,
	)

	rep := reporters.Create()

	if cfg.PushGateway != "" {
		rep.Add(reporters.NewCountriesReporter(cfg.PushGateway))
	}

	if len(cfg.Reports.Countries) > 0 {
		rep.Add(reporters.NewUpdatesReporter(&cfg.Reports, covidDB))
	}

	covidProbe := covidprobe.NewProbe(cfg.RapidAPI.Key, covidDB, rep)
	populationProbe := popprobe.Create(cfg.RapidAPI.Key, popDB)

	for {
		if ok := monitor.Run(&cfg, covidProbe, populationProbe); !ok || cfg.Once {
			break
		}

		time.Sleep(cfg.Interval)
	}
}
