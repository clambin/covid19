package db

import (
	"database/sql"
	"errors"
	"fmt"
)

// PopulationStore represents a database holding population count per country
//
//go:generate mockery --name PopulationStore
type PopulationStore interface {
	List() (map[string]int64, error)
	Add(string, int64) error
}

var _ PopulationStore = &PGPopulationStore{}

// PGPopulationStore implements PopulationStore for Postgres databases
type PGPopulationStore struct {
	DB *DB
}

// NewPopulationStore creates a new PostgresDB object
func NewPopulationStore(db *DB) *PGPopulationStore {
	return &PGPopulationStore{DB: db}
}

// List all records from the population table
func (store *PGPopulationStore) List() (entries map[string]int64, err error) {
	var rows *sql.Rows
	rows, err = store.DB.Handle.Query(`SELECT country_code, population FROM population`)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = nil
		}
		return
	}
	defer func() {
		_ = rows.Close()
	}()

	entries = make(map[string]int64, 0)

	for rows.Next() {
		var code string
		var population int64
		if rows.Scan(&code, &population) == nil {
			entries[code] = population
		}
	}

	return entries, err
}

// Add to population database table. If a record for the specified country code already exists, it will be updated
func (store *PGPopulationStore) Add(code string, pop int64) error {
	sqlStr := fmt.Sprintf(`INSERT INTO population(country_code, population) VALUES ('%s', %d)  ON CONFLICT (country_code) DO UPDATE SET population = EXCLUDED.population`,
		code, pop)

	stmt, err := store.DB.Handle.Prepare(sqlStr)
	if err == nil {
		_, err = stmt.Exec()
	}

	return err
}
