package db_test

import (
	"fmt"
	"github.com/clambin/covid19/configuration"
	"github.com/clambin/covid19/db"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

var (
	DB         *db.DB
	covidStore db.CovidStore
	popStore   db.PopulationStore
)

func TestMain(m *testing.M) {
	pg := configuration.LoadPGEnvironment()

	if !pg.IsValid() {
		fmt.Println("Could not find all CovidDB env variables. Skipping this test")
		return
	}

	var err error
	DB, err = db.New(pg)
	if err != nil {
		fmt.Printf("unable to connect to database: %s", err.Error())
		os.Exit(1)
	}

	covidStore = db.NewCovidStore(DB)
	popStore = db.NewPopulationStore(DB)

	m.Run()

	_ = DB.RemoveAll()
}

func TestDB_Failure(t *testing.T) {
	cfg := configuration.PostgresDB{Host: "127.0.0.1", Port: 5432, Database: "test", User: "test", Password: "test"}
	_, err := db.New(cfg)
	assert.Error(t, err)
}
