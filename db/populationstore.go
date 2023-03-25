package db

import (
	"fmt"
)

// PGPopulationStore implements PopulationStore for Postgres databases
type PGPopulationStore struct {
	DB *DB
}

// NewPopulationStore creates a new PostgresDB object
func NewPopulationStore(db *DB) *PGPopulationStore {
	return &PGPopulationStore{DB: db}
}

// List all records from the Population table
func (store *PGPopulationStore) List() (map[string]int64, error) {
	var rows []struct {
		Code       string
		Population int64
	}
	if err := store.DB.Handle.Select(&rows, `SELECT country_code AS "code", population FROM population`); err != nil {
		return nil, err
	}

	entries := make(map[string]int64)
	for _, row := range rows {
		entries[row.Code] = row.Population
	}
	return entries, nil
}

// Add to Population database table. If a record for the specified country code already exists, it will be updated
func (store *PGPopulationStore) Add(code string, pop int64) error {
	stmt, err := store.DB.Handle.Prepare(fmt.Sprintf(
		`INSERT INTO population(country_code, population) VALUES ('%s', %d) ON CONFLICT (country_code) DO UPDATE SET population = EXCLUDED.population`,
		code, pop,
	))
	if err == nil {
		_, err = stmt.Exec()
	}
	return err
}
