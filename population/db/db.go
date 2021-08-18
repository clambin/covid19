package db

import (
	"database/sql"
	"fmt"
	"github.com/clambin/covid19/db"

	// postgres sql driver
	_ "github.com/lib/pq"
)

// DB interface representing a Population database table
//go:generate mockery --name DB
type DB interface {
	List() (map[string]int64, error)
	Add(string, int64) error
}

// PostgresDB implements DB in Postgres
type PostgresDB struct {
	DB *db.DB
}

// New creates a new PostgresDB object
func New(db *db.DB) (pgdb *PostgresDB, err error) {
	pgdb = &PostgresDB{DB: db}
	err = pgdb.initialize()
	return
}

// List all records from the population table
func (db *PostgresDB) List() (entries map[string]int64, err error) {
	var rows *sql.Rows
	rows, err = db.DB.Handle.Query(fmt.Sprintf("SELECT country_code, population FROM population"))

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
func (db *PostgresDB) Add(code string, population int64) (err error) {
	sqlStr := fmt.Sprintf("INSERT INTO population(country_code, population) VALUES ('%s', %d) "+
		"ON CONFLICT (country_code) DO UPDATE SET population = EXCLUDED.population",
		code, population)

	var stmt *sql.Stmt
	stmt, err = db.DB.Handle.Prepare(sqlStr)
	if err == nil {
		_, err = stmt.Exec()
	}

	return
}

// initialize creates the required tables
func (db *PostgresDB) initialize() (err error) {
	_, err = db.DB.Handle.Exec(`
		CREATE TABLE IF NOT EXISTS population (
			country_code TEXT PRIMARY KEY,
			population NUMERIC
		)
	`)

	return
}
