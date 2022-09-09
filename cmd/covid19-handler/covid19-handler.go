package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/clambin/covid19/backfill"
	"github.com/clambin/covid19/configuration"
	"github.com/clambin/covid19/db"
	"github.com/clambin/covid19/simplejsonserver"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"github.com/xonvanetta/shutdown/pkg/shutdown"
	"net/http"
	"os"
	"time"
)

func main() {
	cfg := configuration.GetConfiguration("covid19-handlers", os.Args)

	if cfg.Debug {
		log.SetLevel(log.DebugLevel)
	}

	s, err := CreateStack(cfg)
	if err != nil {
		log.WithError(err).Fatal("app init failed")
	}
	go s.Run()
	<-shutdown.Chan()
	s.Stop()
}

// Stack groups the different components that make up the application
type Stack struct {
	CovidStore      db.CovidStore
	PopulationStore db.PopulationStore
	HTTPServer      *http.Server
	SkipBackfill    bool
}

// CreateStack creates an application stack for the provided configuration
func CreateStack(cfg *configuration.Configuration) (*Stack, error) {
	dbh, err := db.NewWithConfiguration(cfg.Postgres, prometheus.DefaultRegisterer)
	if err != nil {
		return nil, err
	}
	return CreateStackWithStores(cfg, db.NewCovidStore(dbh), db.NewPopulationStore(dbh)), nil
}

// CreateStackWithStores creates an application stack for the provided configuration and Covid-19 store
func CreateStackWithStores(cfg *configuration.Configuration, covidDB db.CovidStore, populationStore db.PopulationStore) (stack *Stack) {
	stack = &Stack{
		CovidStore:      covidDB,
		PopulationStore: populationStore,
	}

	server := simplejsonserver.MakeServer(stack.CovidStore, stack.PopulationStore)
	r := server.GetRouter()
	r.PathPrefix("/debug/pprof/").Handler(http.DefaultServeMux)
	stack.HTTPServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: r,
	}
	return
}

// Run runs the application stack
func (stack *Stack) Run() {
	if !stack.SkipBackfill {
		stack.loadIfEmpty()
	}

	if err := stack.HTTPServer.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		log.WithError(err).Fatal("unable to start grafana SimpleJson server")
	}
}

// Stop stops the application stack
func (stack *Stack) Stop() {
	_ = stack.HTTPServer.Shutdown(context.Background())
}

func (stack *Stack) loadIfEmpty() {
	_, found, err := stack.CovidStore.GetFirstEntry()

	if err != nil {
		log.WithError(err).Fatal("could not access database")
	}

	if found {
		return
	}

	log.Info("database is empty. backfilling ... ")
	bf := backfill.New(stack.CovidStore)
	go func() {
		start := time.Now()
		if err = bf.Run(); err == nil {
			log.Infof("historic data loaded in %s", time.Since(start).String())
		} else {
			log.WithError(err).Error("failed to populate database")
		}
	}()

}
