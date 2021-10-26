package store

import (
	"database/sql"
	"fmt"
	"github.com/clambin/covid19/db"
	log "github.com/sirupsen/logrus"
)

// PopulationStore represents a database holding population count per country
//go:generate mockery --name PopulationStore
type PopulationStore interface {
	List() (map[string]int64, error)
	Add(string, int64) error
}

var _ PopulationStore = &PGPopulationStore{}

// PGPopulationStore implements PopulationStore for Postgres databases
type PGPopulationStore struct {
	DB *db.DB
}

// New creates a new PostgresDB object
func New(db *db.DB) (pgdb *PGPopulationStore) {
	pgdb = &PGPopulationStore{DB: db}
	err := pgdb.initialize()
	if err != nil {
		log.WithError(err).Fatal("failed to open population database")
	}
	return
}

// List all records from the population table
func (store *PGPopulationStore) List() (entries map[string]int64, err error) {
	var rows *sql.Rows
	rows, err = store.DB.Handle.Query(fmt.Sprintf("SELECT country_code, population FROM population"))

	if err == nil {
		entries = make(map[string]int64, 0)

		for rows.Next() {
			var code string
			var population int64
			if rows.Scan(&code, &population) == nil {
				entries[code] = population
			}
		}

		_ = rows.Close()
	}

	return entries, err
}

// Add to population database table. If a record for the specified country code already exists, it will be updated
func (store *PGPopulationStore) Add(code string, population int64) (err error) {
	sqlStr := fmt.Sprintf("INSERT INTO population(country_code, population) VALUES ('%s', %d) "+
		"ON CONFLICT (country_code) DO UPDATE SET population = EXCLUDED.population",
		code, population)

	var stmt *sql.Stmt
	stmt, err = store.DB.Handle.Prepare(sqlStr)
	if err == nil {
		_, err = stmt.Exec()
	}

	return
}

// initialize creates the required tables
func (store *PGPopulationStore) initialize() (err error) {
	_, err = store.DB.Handle.Exec(`
		CREATE TABLE IF NOT EXISTS population (
			country_code TEXT PRIMARY KEY,
			population NUMERIC
		)
	`)

	return
}