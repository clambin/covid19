package coviddb

import (
	"database/sql"
	"fmt"
	"github.com/clambin/covid19/db"
	"github.com/lib/pq"
	"time"
)

// DB interface representing a Covid Database
type DB interface {
	List(time.Time) ([]CountryEntry, error)
	ListLatestByCountry() (map[string]time.Time, error)
	GetFirstEntry() (time.Time, bool, error)
	GetLastBeforeDate(string, time.Time) (*CountryEntry, bool, error)
	Add([]CountryEntry) error
	GetAllCountryCodes() (codes []string, err error)
}

// PostgresDB implementation of DB
type PostgresDB struct {
	DB *db.DB
}

// New create a new PostgresDB object
func New(db *db.DB) (pgdb *PostgresDB, err error) {
	pgdb = &PostgresDB{DB: db}
	err = pgdb.initialize()
	return
}

// CountryEntry represents one row in the Covid DB
type CountryEntry struct {
	Timestamp time.Time
	Code      string
	Name      string
	Confirmed int64
	Recovered int64
	Deaths    int64
}

// List retrieved all records from the database up to endDate
func (db *PostgresDB) List(endDate time.Time) (entries []CountryEntry, err error) {
	var rows *sql.Rows
	rows, err = db.DB.Handle.Query(fmt.Sprintf(
		"SELECT time, country_code, country_name, confirmed, recovered, death FROM covid19 WHERE time <= '%s' ORDER BY 1",
		endDate.Format("2006-01-02 15:04:05")))

	if err == nil {
		for rows.Next() {
			var entry CountryEntry
			if rows.Scan(&entry.Timestamp, &entry.Code, &entry.Name, &entry.Confirmed, &entry.Recovered, &entry.Deaths) == nil {
				entries = append(entries, entry)
			}
		}

		_ = rows.Close()
	}

	return
}

// ListLatestByCountry returns the timestamp of each country's last update
func (db *PostgresDB) ListLatestByCountry() (entries map[string]time.Time, err error) {
	entries = make(map[string]time.Time)
	var rows *sql.Rows
	rows, err = db.DB.Handle.Query(`SELECT country_name, MAX(time) FROM covid19 GROUP BY country_name`)

	if err == nil {
		for rows.Next() {
			var country string
			var timestamp time.Time

			if rows.Scan(&country, &timestamp) == nil {
				entries[country] = timestamp
			}
		}

		_ = rows.Close()
	}
	return
}

// GetFirstEntry returns the timestamp of the first entry
func (db *PostgresDB) GetFirstEntry() (first time.Time, found bool, err error) {
	err = db.DB.Handle.QueryRow(`SELECT MIN(time) FROM covid19`).Scan(&first)
	found = err == nil

	if err != nil && err.Error() == "sql: Scan error on column index 0, name \"min\": unsupported Scan, storing driver.Value type <nil> into type *time.Time" {
		err = nil
	}

	return
}

// GetLastBeforeDate gets the last entry for a country before a specified data.
func (db *PostgresDB) GetLastBeforeDate(countryName string, before time.Time) (result *CountryEntry, found bool, err error) {
	var last time.Time
	// FIXME: leaving out sprintf gives errors on processing timestamp???
	err = db.DB.Handle.QueryRow(
		fmt.Sprintf("SELECT MAX(time) FROM covid19 WHERE country_name = '%s' AND time < '%s'", countryName, before.Format("2006-01-02 15:04:05"))).Scan(&last)

	// row.Scan() should return sql.ErrNoRows ???
	if err != nil {
		if err.Error() != "sql: Scan error on column index 0, name \"max\": unsupported Scan, storing driver.Value type <nil> into type *time.Time" {
			return nil, false, fmt.Errorf("unable to get coviddb data: %v", err)
		}
		err = nil
	} else {
		result = &CountryEntry{Timestamp: before, Name: countryName}
		err = db.DB.Handle.QueryRow(
			fmt.Sprintf("SELECT country_code, confirmed, death, recovered FROM covid19 where country_name = '%s' and time = '%s'",
				countryName,
				last.Format("2006-01-02 15:04:05")),
		).Scan(&result.Code, &result.Confirmed, &result.Deaths, &result.Recovered)

		found = err == nil
	}

	return
}

// Add inserts all specified records in the covid19 database table
func (db *PostgresDB) Add(entries []CountryEntry) (err error) {
	var tx *sql.Tx
	tx, err = db.DB.Handle.Begin()

	if err == nil {
		var stmt *sql.Stmt
		stmt, err = tx.Prepare(pq.CopyIn(
			"covid19",
			"time", "country_code", "country_name", "confirmed", "death", "recovered",
		))

		if err == nil {
			for _, entry := range entries {
				_, err = stmt.Exec(entry.Timestamp, entry.Code, entry.Name, entry.Confirmed, entry.Deaths, entry.Recovered)
				if err != nil {
					break
				}
			}

			if err == nil {
				_, err = stmt.Exec()
			}

			if err == nil {
				err = tx.Commit()
			}

			_ = stmt.Close()
		}
	}

	return err
}

// RemoveDB removes all tables & indexes
func (db *PostgresDB) RemoveDB() (err error) {
	_, err = db.DB.Handle.Exec(`DROP TABLE IF EXISTS covid19 CASCADE`)
	return
}

// GetAllCountryCodes retrieves all country codes present in the covid table
func (db *PostgresDB) GetAllCountryCodes() (codes []string, err error) {
	var rows *sql.Rows
	rows, err = db.DB.Handle.Query(`SELECT distinct country_code FROM covid19`)

	if err == nil {
		for rows.Next() {
			var code string
			if rows.Scan(&code) == nil {
				codes = append(codes, code)
			}
		}

		_ = rows.Close()

	}

	return
}

// initialize opens a db connection and created the required tables
func (db *PostgresDB) initialize() (err error) {
	_, err = db.DB.Handle.Exec(`
		CREATE TABLE IF NOT EXISTS covid19 (
			time TIMESTAMP WITHOUT TIME ZONE,
			country_code TEXT,
			country_name TEXT,
			confirmed BIGINT,
			death BIGINT,
			recovered BIGINT
		);
		CREATE INDEX IF NOT EXISTS idx_covid_country_name ON covid19(country_name);
		CREATE INDEX IF NOT EXISTS idx_covid_country_code ON covid19(country_code);
		CREATE INDEX IF NOT EXISTS idx_covid_time ON covid19(time);
	`)

	return
}
