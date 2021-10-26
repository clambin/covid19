package main

import (
	"context"
	"fmt"
	"github.com/clambin/covid19/cache"
	"github.com/clambin/covid19/configuration"
	covidStore "github.com/clambin/covid19/covid/store"
	"github.com/clambin/covid19/db"
	"github.com/clambin/covid19/handler"
	grafana_json "github.com/clambin/grafana-json"
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
	CovidStore covidStore.CovidStore
	Cache      *cache.Cache
	HTTPServer *http.Server
}

// CreateStack creates an application stack for the provided configuration
func CreateStack(cfg *configuration.Configuration) (stack *Stack) {
	dbh, err := db.NewWithConfiguration(&cfg.Postgres)
	if err != nil {
		panic(err)
	}
	return CreateStackWithStore(cfg, covidStore.New(dbh))
}

// CreateStackWithStore creates an application stack for the provided configuration and Covid-19 store
func CreateStackWithStore(cfg *configuration.Configuration, store covidStore.CovidStore) (stack *Stack) {
	stack = &Stack{
		CovidStore: store,
		Cache:      &cache.Cache{DB: store, Retention: 20 * time.Minute},
	}

	server := &grafana_json.Server{Handler: &handler.Handler{Cache: stack.Cache}}
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
	err := stack.HTTPServer.ListenAndServe()
	if err != http.ErrServerClosed {
		log.WithError(err).Fatal("unable to start grafana SimpleJson server")
	}
}

// Stop stops the applicayion stack
func (stack *Stack) Stop() {
	_ = stack.HTTPServer.Shutdown(context.Background())
}