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
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"net/http"
	"sync"
	"time"
)

// Stack groups the different components that make up the application
type Stack struct {
	Cfg             *configuration.Configuration
	CovidStore      db.CovidStore
	PopulationStore db.PopulationStore
	Servers         HTTPServers
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
func CreateStackWithStores(cfg *configuration.Configuration, covidStore db.CovidStore, populationStore db.PopulationStore) (stack *Stack) {
	m := http.NewServeMux()
	m.Handle("/metrics", promhttp.Handler())

	server := simplejsonserver.MakeServer(covidStore, populationStore)
	r := server.GetRouter()

	return &Stack{
		Cfg:             cfg,
		CovidStore:      covidStore,
		PopulationStore: populationStore,
		Servers: HTTPServers{
			"prometheus": &http.Server{Addr: fmt.Sprintf(":%d", cfg.PrometheusPort), Handler: m},
			"SimpleJSON": &http.Server{Addr: fmt.Sprintf(":%d", cfg.Port), Handler: r},
		},
	}
}

// RunHandler runs the SimpleJSON server
func (stack *Stack) RunHandler() {
	wg := sync.WaitGroup{}
	stack.Servers.Start(&wg)
	wg.Wait()
}

// StopHandler stops the SimpleJSON server
func (stack *Stack) StopHandler() {
	stack.Servers.Stop(5 * time.Second)
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
	cp := populationProbe.New(stack.Cfg.Monitor.RapidAPIKey, stack.PopulationStore)
	if count, err := cp.Update(context.Background()); err == nil {
		log.Infof("discovered %d country population figures in %v", count, time.Since(start))
	} else {
		log.WithError(err).Error("failed to update population figures")
	}
}
