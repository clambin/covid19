package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/clambin/covid19/backfill"
	"github.com/clambin/covid19/cache"
	"github.com/clambin/covid19/configuration"
	covidStore "github.com/clambin/covid19/covid/store"
	"github.com/clambin/covid19/db"
	"github.com/clambin/covid19/handler"
	populationStore "github.com/clambin/covid19/population/store"
	"github.com/clambin/simplejson"
	log "github.com/sirupsen/logrus"
	"github.com/xonvanetta/shutdown/pkg/shutdown"
	"net/http"
	"os"
	"time"
)

func main() {
	cfg := configuration.GetConfiguration("covid19-handler", os.Args)

	if cfg.Debug {
		log.SetLevel(log.DebugLevel)
	}

	s := CreateStack(cfg)
	go s.Run()
	<-shutdown.Chan()
	s.Stop()
}

// Stack groups the different components that make up the application
type Stack struct {
	CovidStore      covidStore.CovidStore
	Cache           *cache.Cache
	PopulationStore populationStore.PopulationStore
	HTTPServer      *http.Server
	SkipBackfill    bool
}

// CreateStack creates an application stack for the provided configuration
func CreateStack(cfg *configuration.Configuration) (stack *Stack) {
	dbh, err := db.NewWithConfiguration(cfg.Postgres)
	if err != nil {
		panic(err)
	}
	return CreateStackWithStores(cfg, covidStore.New(dbh), populationStore.New(dbh))
}

// CreateStackWithStores creates an application stack for the provided configuration and Covid-19 store
func CreateStackWithStores(cfg *configuration.Configuration, covidDB covidStore.CovidStore, populationStore populationStore.PopulationStore) (stack *Stack) {
	stack = &Stack{
		CovidStore:      covidDB,
		PopulationStore: populationStore,
		Cache:           &cache.Cache{DB: covidDB, Retention: 20 * time.Minute},
	}

	server := &simplejson.Server{
		Name: "covid19",
		Handlers: []simplejson.Handler{
			&handler.Handler{
				Cache:           stack.Cache,
				PopulationStore: populationStore,
			},
		},
	}
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
	if stack.SkipBackfill == false {
		stack.loadIfEmpty()
	}

	if err := stack.HTTPServer.ListenAndServe(); errors.Is(err, http.ErrServerClosed) == false {
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

	if found == false {
		log.Info("database is empty. backfilling ... ")
		bf := backfill.New(stack.CovidStore)
		go func() {
			start := time.Now()
			if err = bf.Run(); err == nil {
				log.Infof("historic data loaded in %s", time.Now().Sub(start).String())
			} else {
				log.WithError(err).Error("failed to populate database")
			}
		}()
	}
}
