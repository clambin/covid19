package db

import (
	"database/sql"
	"fmt"
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
	db = &DB{
		psqlInfo: fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			host, port, user, password, database),
		database: database,
	}

	db.Handle, err = sql.Open("postgres", db.psqlInfo)

	if err == nil {
		prometheus.MustRegister(collectors.NewDBStatsCollector(db.Handle, db.database))
	}

	return
}
