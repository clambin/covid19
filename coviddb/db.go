package coviddb

import (
	"database/sql"
	"fmt"
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
	psqlInfo    string
	initialized bool
}

// NewPostgresDB create a new PostgresDB object
func NewPostgresDB(host string, port int, database string, user string, password string) *PostgresDB {
	return &PostgresDB{
		psqlInfo: fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			host, port, user, password, database),
		initialized: false,
	}
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

func (db *PostgresDB) openDB() (dbh *sql.DB, err error) {
	dbh, err = sql.Open("postgres", db.psqlInfo)
	if err != nil {
		return nil, fmt.Errorf("could not connect to covid DB: %s", err.Error())
	}

	err = db.initializeDB(dbh)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize covid db: %s", err.Error())
	}

	return
}

// List retrieved all records from the database up to endDate
func (db *PostgresDB) List(endDate time.Time) (entries []CountryEntry, err error) {
	var dbh *sql.DB
	dbh, err = db.openDB()
	if err == nil {
		var rows *sql.Rows
		rows, err = dbh.Query(fmt.Sprintf(
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
		_ = dbh.Close()
	}
	return entries, err
}

// ListLatestByCountry returns the timestamp of each country's last update
func (db *PostgresDB) ListLatestByCountry() (entries map[string]time.Time, err error) {
	var dbh *sql.DB
	dbh, err = db.openDB()
	if err == nil {
		entries = make(map[string]time.Time)
		var rows *sql.Rows
		if rows, err = dbh.Query(`SELECT country_name, MAX(time) FROM covid19 GROUP BY country_name`); err == nil {
			for rows.Next() {
				var (
					country   string
					timestamp time.Time
				)
				if rows.Scan(&country, &timestamp) == nil {
					entries[country] = timestamp
				}
			}
			_ = rows.Close()
		}
		_ = dbh.Close()
	}
	return entries, err
}

// GetFirstEntry returns the timestamp of the first entry
func (db *PostgresDB) GetFirstEntry() (first time.Time, found bool, err error) {
	var dbh *sql.DB
	dbh, err = db.openDB()
	if err == nil {
		err = dbh.QueryRow(`SELECT MIN(time) FROM covid19`).Scan(&first)
		found = err == nil
		_ = dbh.Close()
	}
	return
}

// GetLastBeforeDate gets the last entry for a country before a specified data.
func (db *PostgresDB) GetLastBeforeDate(countryName string, before time.Time) (result *CountryEntry, found bool, err error) {
	var dbh *sql.DB
	dbh, err = db.openDB()
	if err == nil {
		var last time.Time
		// FIXME: leaving out sprintf gives errors on processing timestamp???
		err = dbh.QueryRow(fmt.Sprintf(
			"SELECT MAX(time) FROM covid19 WHERE country_name = '%s' AND time < '%s'", countryName, before.Format("2006-01-02 15:04:05"),
		)).Scan(&last)

		// row.Scan() should return sql.ErrNoRows ???
		if err != nil && err.Error() == "sql: Scan error on column index 0, name \"max\": unsupported Scan, storing driver.Value type <nil> into type *time.Time" {
			err = nil
		} else if err == nil {
			result = &CountryEntry{Timestamp: before, Name: countryName}
			err = dbh.QueryRow(
				fmt.Sprintf(
					"SELECT country_code, confirmed, death, recovered FROM covid19 where country_name = '%s' and time = '%s'",
					countryName,
					last.Format("2006-01-02 15:04:05"),
				),
			).Scan(&result.Code, &result.Confirmed, &result.Deaths, &result.Recovered)

			found = err == nil
		}
		_ = dbh.Close()
	}
	return
}

// Add inserts all specified records in the covid19 database table
func (db *PostgresDB) Add(entries []CountryEntry) (err error) {
	var dbh *sql.DB
	dbh, err = db.openDB()
	if err != nil {
		return
	}

	defer func() {
		_ = dbh.Close()
	}()

	var tx *sql.Tx
	tx, err = dbh.Begin()

	if err != nil {
		return fmt.Errorf("failed to start transaction for coviddb: %s", err.Error())
	}

	var stmt *sql.Stmt
	stmt, err = tx.Prepare(pq.CopyIn(
		"covid19",
		"time", "country_code", "country_name", "confirmed", "death", "recovered",
	))

	if err != nil {
		return fmt.Errorf("failed to add to prepare statement for coviddb: %s", err.Error())
	}

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

	return err
}

// initializeDB created the required tables
func (db *PostgresDB) initializeDB(dbh *sql.DB) (err error) {
	if db.initialized {
		return
	}

	_, err = dbh.Exec(`
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

	db.initialized = err == nil

	return err
}

// RemoveDB removes all tables & indexes
func (db *PostgresDB) RemoveDB() (err error) {
	var dbh *sql.DB
	dbh, err = sql.Open("postgres", db.psqlInfo)
	if err == nil {
		_, err = dbh.Exec(`DROP TABLE IF EXISTS covid19 CASCADE`)
		db.initialized = err != nil
		_ = dbh.Close()
	}
	return
}

func (db *PostgresDB) GetAllCountryCodes() (codes []string, err error) {
	var dbh *sql.DB
	dbh, err = db.openDB()
	if err == nil {
		var rows *sql.Rows
		rows, err = dbh.Query(`SELECT distinct country_code FROM covid19`)

		if err == nil {
			for rows.Next() {
				var code string
				if rows.Scan(&code) == nil {
					codes = append(codes, code)
				}
			}
			_ = rows.Close()
		}
		_ = dbh.Close()
	}
	return
}
