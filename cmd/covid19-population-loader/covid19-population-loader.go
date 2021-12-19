package main

import (
	"context"
	"github.com/clambin/covid19/configuration"
	"github.com/clambin/covid19/db"
	populationProbe "github.com/clambin/covid19/population/probe"
	populationStore "github.com/clambin/covid19/population/store"
	log "github.com/sirupsen/logrus"
	"os"
)

func main() {
	cfg := configuration.GetConfiguration("covid19-population-loader", os.Args)

	if cfg.Debug {
		log.SetLevel(log.DebugLevel)
	}

	dbh, err := db.NewWithConfiguration(cfg.Postgres)
	if err != nil {
		panic(err)
	}

	ps := populationStore.New(dbh)
	cp := populationProbe.New(cfg.Monitor.RapidAPIKey.Get(), ps)

	err = cp.Update(context.Background())

	if err != nil {
		log.WithError(err).Error("failed to update population figures")
		os.Exit(1)
	}
}
