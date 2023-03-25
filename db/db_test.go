package db_test

import (
	"bytes"
	"fmt"
	"github.com/clambin/covid19/configuration"
	"github.com/clambin/covid19/db"
	"github.com/stretchr/testify/assert"
	"testing"
)

var (
	DB         *db.DB
	covidStore *db.PGCovidStore
	popStore   *db.PGPopulationStore
)

func TestMain(m *testing.M) {
	cfg := `
postgres:
  host: "$pg_host"
  port: $pg_port
  database: "$pg_database"
  user: "$pg_user"
  password: "$pg_password"
`
	config, err := configuration.LoadConfiguration(bytes.NewBufferString(cfg))
	if err != nil {
		panic(err)
	}

	if !config.Postgres.IsValid() {
		fmt.Println("Could not find all DB env variables. Skipping this test")
		return
	}

	if DB, err = db.New(config.Postgres); err != nil {
		panic(fmt.Errorf("unable to connect to database: %w", err))
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
