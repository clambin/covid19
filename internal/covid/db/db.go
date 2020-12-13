package db

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/lib/pq"
	log "github.com/sirupsen/logrus"
)

// DB interface representing a Covid Database
type DB interface {
	List(time.Time) ([]CountryEntry, error)
	ListLatestByCountry() (map[string]time.Time, error)
	GetFirstEntry() (time.Time, error)
	Add([]CountryEntry) error
}

// PostgresDB implementation of DB
type PostgresDB struct {
	psqlInfo    string
	initialized bool
}

// NewPostgresDB create a new PostgresDB object
func NewPostgresDB(host string, port int, database string, user string, password string) *PostgresDB {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, database)

	return &PostgresDB{psqlInfo: psqlInfo, initialized: false}
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
func (db *PostgresDB) List(endDate time.Time) ([]CountryEntry, error) {
	entries := make([]CountryEntry, 0)

	dbh, err := sql.Open("postgres", db.psqlInfo)

	if err == nil {
		defer dbh.Close()
		err = db.initializeDB(dbh)
	}

	if err == nil {
		rows, err := dbh.Query(fmt.Sprintf(
			"SELECT time, country_code, country_name, confirmed, recovered, death FROM covid19 WHERE time <= '%s' ORDER BY 1",
			endDate.Format("2006-01-02 15:04:05")))

		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var entry CountryEntry
				err = rows.Scan(&entry.Timestamp, &entry.Code, &entry.Name, &entry.Confirmed, &entry.Recovered, &entry.Deaths)
				if err != nil {
					break
				}
				entries = append(entries, entry)
			}
			log.Debugf("Found %d records", len(entries))
		}
	}

	return entries, err
}

// ListLatestByCountry returns the timestamp of each country's last update
func (db *PostgresDB) ListLatestByCountry() (map[string]time.Time, error) {
	entries := make(map[string]time.Time, 0)

	dbh, err := sql.Open("postgres", db.psqlInfo)

	if err == nil {
		defer dbh.Close()
		err = db.initializeDB(dbh)
	}

	if err == nil {
		rows, err := dbh.Query(`SELECT country_name, MAX(time) FROM covid19 GROUP BY country_name`)

		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var (
					country   string
					timestamp time.Time
				)
				err = rows.Scan(&country, &timestamp)
				if err != nil {
					break
				}
				entries[country] = timestamp
			}
			log.Debugf("Found %d records", len(entries))
		}
	}

	return entries, err
}

// GetFirstEntry returns the timestamp of the first entry
func (db *PostgresDB) GetFirstEntry() (time.Time, error) {
	first := time.Time{}

	dbh, err := sql.Open("postgres", db.psqlInfo)

	if err == nil {
		defer dbh.Close()
		err = db.initializeDB(dbh)
	}

	if err == nil {
		rows, err := dbh.Query(`SELECT MIN(time) FROM covid19`)

		if err == nil {
			defer rows.Close()
			for rows.Next() {
				err = rows.Scan(&first)
				if err != nil {
					log.Debug(err)
					break
				}
			}
			log.Debugf("First record: %s", first.String())
		}
	}

	return first, err
}

// Add inserts all specified records in the covid19 database table
func (db *PostgresDB) Add(entries []CountryEntry) error {

	dbh, err := sql.Open("postgres", db.psqlInfo)

	if err == nil {
		defer dbh.Close()
		err = db.initializeDB(dbh)
	}

	if err == nil {
		txn, err := dbh.Begin()
		if err != nil {
			return err
		}

		stmt, err := txn.Prepare(pq.CopyIn("covid19", "time", "country_code", "country_name", "confirmed", "death", "recovered"))
		if err != nil {
			return err
		}

		for _, entry := range entries {
			_, err = stmt.Exec(entry.Timestamp, entry.Code, entry.Name, entry.Confirmed, entry.Deaths, entry.Recovered)
			if err != nil {
				return err
			}
		}

		_, err = stmt.Exec()
		if err != nil {
			return err
		}

		err = stmt.Close()
		if err != nil {
			return err
		}

		err = txn.Commit()
	}
	return nil
}

// initializeDB created the required tables
func (db *PostgresDB) initializeDB(dbh *sql.DB) error {
	if db.initialized {
		return nil
	}

	_, err := dbh.Exec(`
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

	if err == nil {
		db.initialized = true
	}

	return err
}

// RemoveDB removes all tables & indexes
func (db *PostgresDB) RemoveDB() error {

	dbh, err := sql.Open("postgres", db.psqlInfo)

	if err == nil {
		defer dbh.Close()

		_, err = dbh.Exec(`DROP TABLE IF EXISTS covid19 CASCADE`)

		if err == nil {
			db.initialized = false
		}
	}

	return err
}
