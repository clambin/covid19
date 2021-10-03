package application

import (
	"context"
	"fmt"
	"github.com/clambin/covid19/cache"
	"github.com/clambin/covid19/configuration"
	covidProbe "github.com/clambin/covid19/covid/probe"
	covidStore "github.com/clambin/covid19/covid/store"
	"github.com/clambin/covid19/db"
	"github.com/clambin/covid19/handler"
	populationProbe "github.com/clambin/covid19/population/probe"
	populationStore "github.com/clambin/covid19/population/store"
	grafana_json "github.com/clambin/grafana-json"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"net/http"
	"time"
)

// Stack groups the different components that make up the application
type Stack struct {
	Cache              *cache.Cache
	CovidStore         covidStore.CovidStore
	CovidProbe         *covidProbe.Covid19Probe
	CovidProbeInterval time.Duration
	PopulationStore    populationStore.PopulationStore
	PopulationProbe    *populationProbe.Probe
	HTTPServer         *http.Server
}

// New creates a new application based for the provided configuration
func New(cfg *configuration.Configuration) *Stack {
	covidDB, populationDB := openDatabases(cfg)
	return NewWithDatabases(cfg, covidDB, populationDB)
}

// NewWithDatabases creates a new application based for the provided configuration and databases
func NewWithDatabases(cfg *configuration.Configuration, covidDB covidStore.CovidStore, populationDB populationStore.PopulationStore) (stack *Stack) {
	c := &cache.Cache{DB: covidDB, Retention: cfg.Monitor.Interval}

	server := &grafana_json.Server{Handler: &handler.Handler{Cache: c}}
	r := server.GetRouter()
	r.PathPrefix("/debug/pprof/").Handler(http.DefaultServeMux)

	stack = &Stack{
		Cache:              c,
		CovidStore:         covidDB,
		CovidProbe:         covidProbe.New(&cfg.Monitor, covidDB),
		CovidProbeInterval: cfg.Monitor.Interval,
		PopulationStore:    populationDB,
		PopulationProbe:    populationProbe.New(cfg.Monitor.RapidAPIKey.Value, populationDB),
		HTTPServer: &http.Server{
			Addr:    fmt.Sprintf(":%d", cfg.Port),
			Handler: r,
		},
	}
	prometheus.MustRegister(stack.CovidProbe)
	return
}

func openDatabases(cfg *configuration.Configuration) (covidDB covidStore.CovidStore, populationDB populationStore.PopulationStore) {
	DB, err := db.New(
		cfg.Postgres.Host,
		cfg.Postgres.Port,
		cfg.Postgres.Database,
		cfg.Postgres.User,
		cfg.Postgres.Password,
	)

	if err != nil {
		log.WithError(err).Fatalf("unable to access probe DB '%s'", cfg.Postgres.Database)
	}

	covidDB = covidStore.New(DB)
	populationDB = populationStore.New(DB)

	return
}

// Run runs the application stack.  Cancel the provided context to terminate the stack
func (stack *Stack) Run(ctx context.Context) {
	go func() {
		err := stack.HTTPServer.ListenAndServe()
		if err != http.ErrServerClosed {
			log.WithError(err).Fatal("unable to start grafana SimpleJson server")
		}
	}()

	go stack.PopulationProbe.Update(ctx)

	populationTicker := time.NewTicker(24 * time.Hour)
	covidTicker := time.NewTicker(stack.CovidProbeInterval)

	for running := true; running; {
		select {
		case <-ctx.Done():
			running = false
		case <-populationTicker.C:
			stack.PopulationProbe.Update(ctx)
		case <-covidTicker.C:
			err := stack.CovidProbe.Update(ctx)
			if err != nil {
				log.WithError(err).Error("failed to get COVID-19 statistics")
			}
		}
	}

	_ = stack.HTTPServer.Shutdown(context.Background())
	covidTicker.Stop()
	populationTicker.Stop()
}
