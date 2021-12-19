package main

import (
	"context"
	"github.com/clambin/covid19/configuration"
	covidProbe "github.com/clambin/covid19/covid/probe"
	covidStore "github.com/clambin/covid19/covid/store"
	"github.com/clambin/covid19/db"
	log "github.com/sirupsen/logrus"
	"os"
)

func main() {
	cfg := configuration.GetConfiguration("covid19-loader", os.Args)

	if cfg.Debug {
		log.SetLevel(log.DebugLevel)
	}

	dbh, err := db.NewWithConfiguration(cfg.Postgres)
	if err != nil {
		panic(err)
	}

	store := covidStore.New(dbh)
	cp := covidProbe.New(&cfg.Monitor, store)

	err = cp.Update(context.Background())

	if err != nil {
		log.WithError(err).Error("failed to update COVID-19 figures")
		os.Exit(1)
	}
}
