package db

import (
	"database/sql"
	"fmt"
	"github.com/clambin/covid19/configuration"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	// postgres sql driver
	_ "github.com/lib/pq"
)

// DB hold the handle to a database.  Provides a Prometheus DBStatsCollector to monitor DB connections
type DB struct {
	Handle   *sql.DB
	psqlInfo string
	database string
}

// New created a new DB object and connects to the database
func New(host string, port int, database string, user string, password string) (db *DB, err error) {
	cfg := &configuration.PostgresDB{
		Host:     host,
		Port:     port,
		Database: database,
		User:     user,
		Password: password,
	}
	return NewWithConfiguration(cfg)
}

// NewWithConfiguration creates a new DB connector
func NewWithConfiguration(cfg *configuration.PostgresDB) (db *DB, err error) {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Database)
	db = &DB{
		psqlInfo: psqlInfo,
		database: cfg.Database,
	}
	db.Handle, err = sql.Open("postgres", db.psqlInfo)
	prometheus.MustRegister(collectors.NewDBStatsCollector(db.Handle, db.database))
	return
}
