package stack

import (
	"context"
	"fmt"
	"github.com/clambin/covid19/backfill"
	"github.com/clambin/covid19/configuration"
	covidProbe "github.com/clambin/covid19/covid"
	"github.com/clambin/covid19/db"
	populationProbe "github.com/clambin/covid19/population"
	"github.com/clambin/covid19/simplejsonserver"
	"github.com/clambin/simplejson/v6"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/exp/slog"
	"net/http"
	"time"
)

// Stack groups the different components that make up the application
type Stack struct {
	Cfg              *configuration.Configuration
	DB               *db.DB
	CovidStore       db.CovidStore
	PopulationStore  db.PopulationStore
	SimpleJSONServer *simplejson.Server
}

var _ prometheus.Collector = &Stack{}

// CreateStack creates an application stack for the provided configuration
func CreateStack(cfg *configuration.Configuration) (*Stack, error) {
	dbh, err := db.New(cfg.Postgres)
	if err != nil {
		return nil, fmt.Errorf("database: %w", err)
	}

	covidStore := db.NewCovidStore(dbh)
	populationStore := db.NewPopulationStore(dbh)

	return &Stack{
		Cfg:              cfg,
		DB:               dbh,
		CovidStore:       covidStore,
		PopulationStore:  populationStore,
		SimpleJSONServer: simplejsonserver.New(covidStore, populationStore),
	}, nil
}

// RunHandler runs the SimpleJSON server
func (stack *Stack) RunHandler() error {
	return http.ListenAndServe(fmt.Sprintf(":%d", stack.Cfg.Port), stack.SimpleJSONServer)
}

// Load retrieves the latest covid19 figures and stores them in the database
func (stack *Stack) Load() {
	if stack.loadIfEmpty() {
		return
	}

	start := time.Now()
	cp := covidProbe.New(&stack.Cfg.Monitor, stack.CovidStore)
	if count, err := cp.Update(context.Background()); err == nil {
		slog.Info("discovered country population figures", "count", count, "duration", time.Since(start))
	} else {
		slog.Error("failed to update COVID-19 figures", err)
	}
}

func (stack *Stack) loadIfEmpty() bool {
	if rows, err := stack.CovidStore.Rows(); err != nil {
		slog.Error("could not access database", err)
		return false
	} else if rows > 0 {
		return false
	}

	slog.Info("database is empty. backfilling ... ")

	start := time.Now()
	bf := backfill.New(stack.CovidStore)
	if err := bf.Run(); err != nil {
		slog.Error("failed to populate database", err)
		return false
	}

	slog.Info("historic data loaded", "duration", time.Since(start))
	return true
}

// LoadPopulation retrieves the latest population figures and stores them in the database
func (stack *Stack) LoadPopulation() {
	start := time.Now()
	cp := populationProbe.New(stack.Cfg.Monitor.RapidAPIKey, stack.PopulationStore)
	if count, err := cp.Update(context.Background()); err == nil {
		slog.Info("discovered country population figures", "count", count, "duration", time.Since(start))
	} else {
		slog.Error("failed to update population figures", err)
	}
}

// Describe implements the prometheus.Collector interface
func (stack *Stack) Describe(descs chan<- *prometheus.Desc) {
	stack.DB.Collector.Describe(descs)
	stack.SimpleJSONServer.Describe(descs)
}

// Collect implements the prometheus.Collector interface
func (stack *Stack) Collect(metrics chan<- prometheus.Metric) {
	stack.DB.Collector.Collect(metrics)
	stack.SimpleJSONServer.Collect(metrics)
}
