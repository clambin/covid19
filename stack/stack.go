package stack

import (
	"context"
	"errors"
	"fmt"
	"github.com/clambin/covid19/backfill"
	"github.com/clambin/covid19/configuration"
	covidProbe "github.com/clambin/covid19/covid"
	"github.com/clambin/covid19/db"
	populationProbe "github.com/clambin/covid19/population"
	"github.com/clambin/covid19/simplejsonserver"
	log "github.com/sirupsen/logrus"
	"net/http"
	"time"
)

// Stack groups the different components that make up the application
type Stack struct {
	Cfg             *configuration.Configuration
	CovidStore      db.CovidStore
	PopulationStore db.PopulationStore
	HTTPServer      *http.Server
	SkipBackFill    bool
}

// CreateStack creates an application stack for the provided configuration
func CreateStack(cfg *configuration.Configuration) (*Stack, error) {
	dbh, err := db.New(cfg.Postgres)
	if err != nil {
		log.WithError(err).Fatal("failed to connect to database")
	}
	return CreateStackWithStores(cfg, db.NewCovidStore(dbh), db.NewPopulationStore(dbh)), nil
}

// CreateStackWithStores creates an application stack for the provided configuration and stores
func CreateStackWithStores(cfg *configuration.Configuration, covidDB db.CovidStore, populationStore db.PopulationStore) (stack *Stack) {
	return &Stack{
		Cfg:             cfg,
		CovidStore:      covidDB,
		PopulationStore: populationStore,
	}
}

// RunHandler runs the simplejson server
func (stack *Stack) RunHandler() {
	server := simplejsonserver.MakeServer(stack.CovidStore, stack.PopulationStore)
	r := server.GetRouter()
	r.PathPrefix("/debug/pprof/").Handler(http.DefaultServeMux)
	stack.HTTPServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", stack.Cfg.Port),
		Handler: r,
	}
	if err := stack.HTTPServer.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		log.WithError(err).Fatal("unable to start grafana SimpleJson server")
	}
}

// StopHandler stops the simplejson server
func (stack *Stack) StopHandler() {
	_ = stack.HTTPServer.Shutdown(context.Background())
}

// Load retrieves the latest covid19 figures and stores them in the database
func (stack *Stack) Load() {
	if !stack.SkipBackFill {
		if stack.loadIfEmpty() {
			return
		}
	}

	start := time.Now()
	cp := covidProbe.New(&stack.Cfg.Monitor, stack.CovidStore)
	if count, err := cp.Update(context.Background()); err == nil {
		log.Infof("discovered %d country population figures in %v", count, time.Since(start))
	} else {
		log.WithError(err).Error("failed to update COVID-19 figures")
	}
}

func (stack *Stack) loadIfEmpty() bool {
	if _, found, err := stack.CovidStore.GetFirstEntry(); err != nil {
		log.WithError(err).Fatal("could not access database")
	} else if found {
		return false
	}

	log.Info("database is empty. backfilling ... ")

	start := time.Now()
	bf := backfill.New(stack.CovidStore)
	if err := bf.Run(); err != nil {
		log.WithError(err).Error("failed to populate database")
		return false
	}

	log.Infof("historic data loaded in %v", time.Since(start))
	return true
}

// LoadPopulation retrieves the latest population figures and stores them in the database
func (stack *Stack) LoadPopulation() {
	start := time.Now()
	cp := populationProbe.New(stack.Cfg.Monitor.RapidAPIKey.Get(), stack.PopulationStore)
	if count, err := cp.Update(context.Background()); err == nil {
		log.Infof("discovered %d country population figures in %v", count, time.Since(start))
	} else {
		log.WithError(err).Error("failed to update population figures")
	}
}
