package db

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
)

// DB interface representing a Population database table
type DB interface {
	List() (map[string]int64, error)
	Add(string, int64) error
}

// PostgresDB implements DB in Postgres
type PostgresDB struct {
	psqlInfo    string
	initialized bool
}

// NewPostgresDB creates a new PostgresDB object
func NewPostgresDB(host string, port int, database string, user string, password string) *PostgresDB {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, database)

	return &PostgresDB{psqlInfo: psqlInfo, initialized: false}
}

// List all records from the population table
func (db *PostgresDB) List() (map[string]int64, error) {
	var (
		err     error
		dbh     *sql.DB
		rows    *sql.Rows
		entries = make(map[string]int64, 0)
	)

	if err = db.initializeDB(); err == nil {
		if dbh, err = sql.Open("postgres", db.psqlInfo); err == nil {
			rows, err = dbh.Query(fmt.Sprintf("SELECT country_code, population FROM population"))

			if err == nil {
				for rows.Next() {
					var code string
					var population int64
					if rows.Scan(&code, &population) == nil {
						entries[code] = population
					}
				}
				log.Debugf("Found %d records", len(entries))
				_ = rows.Close()
			}
			_ = dbh.Close()
		}
	}
	return entries, err
}

// Add to population database table. If a record for the specified country code already exists, it will be updated
func (db *PostgresDB) Add(code string, population int64) (err error) {
	err = db.initializeDB()

	if err != nil {
		return
	}

	var dbh *sql.DB
	dbh, err = sql.Open("postgres", db.psqlInfo)

	if err != nil {
		return
	}

	// Prepare the SQL statement
	sqlStr := fmt.Sprintf("INSERT INTO population(country_code, population) VALUES ('%s', %d) "+
		"ON CONFLICT (country_code) DO UPDATE SET population = EXCLUDED.population",
		code, population)

	var stmt *sql.Stmt
	stmt, err = dbh.Prepare(sqlStr)
	if err == nil {
		_, err = stmt.Exec()
	} else {
		log.WithError(err).WithField("sql", sqlStr).Error("failed to insert")
	}
	_ = dbh.Close()

	return err
}

// initializeDB creates the required tables
func (db *PostgresDB) initializeDB() (err error) {
	if db.initialized {
		return nil
	}

	var dbh *sql.DB
	dbh, err = sql.Open("postgres", db.psqlInfo)

	if err != nil {
		return
	}

	_, err = dbh.Exec(`
		CREATE TABLE IF NOT EXISTS population (
			country_code TEXT PRIMARY KEY,
			population NUMERIC
		)
	`)

	db.initialized = err == nil
	_ = dbh.Close()

	return err
}
