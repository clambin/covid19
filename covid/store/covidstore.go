package store

import (
	"database/sql"
	"fmt"
	"github.com/clambin/covid19/db"
	"github.com/clambin/covid19/models"
	"github.com/lib/pq"
	log "github.com/sirupsen/logrus"
	"time"
)

// CovidStore represents a database holding COVID-19 statistics
//go:generate mockery --name CovidStore
type CovidStore interface {
	GetAll() (entries []models.CountryEntry, err error)
	GetAllForCountryName(name string) (entries []models.CountryEntry, err error)
	GetLatestForCountries(countryNames []string) (entries map[string]models.CountryEntry, err error)
	GetLatestForCountriesByTime(countryNames []string, endTime time.Time) (entries map[string]models.CountryEntry, err error)
	GetFirstEntry() (first time.Time, found bool, err error)
	Add(entries []models.CountryEntry) (err error)
	GetAllCountryNames() (names []string, err error)
}

var _ CovidStore = &PGCovidStore{}

// PGCovidStore implements CovidStore for Postgres databases
type PGCovidStore struct {
	DB *db.DB
}

// New creates a new PGCovidStore and initializes the database, if necessary
func New(db *db.DB) (store *PGCovidStore) {
	store = &PGCovidStore{DB: db}

	if err := store.initialize(); err != nil {
		log.WithError(err).Fatal("failed to open covid19 database")
	}
	return store
}

// GetAll returns all entries in the database, sorted by timestamp
func (store *PGCovidStore) GetAll() (entries []models.CountryEntry, err error) {
	var rows *sql.Rows
	rows, err = store.DB.Handle.Query(`SELECT time, country_code, country_name, confirmed, recovered, death FROM covid19 ORDER BY 1`)

	if err == nil {
		for rows.Next() {
			var entry models.CountryEntry
			if rows.Scan(&entry.Timestamp, &entry.Code, &entry.Name, &entry.Confirmed, &entry.Recovered, &entry.Deaths) == nil {
				entries = append(entries, entry)
			}
		}
		_ = rows.Close()
	}
	return
}

// GetAllForCountryName returns all entries in the database, sorted by timestamp
func (store *PGCovidStore) GetAllForCountryName(countryName string) (entries []models.CountryEntry, err error) {
	escapedCountryName := EscapeString(countryName)
	var rows *sql.Rows
	rows, err = store.DB.Handle.Query(
		fmt.Sprintf(`SELECT time, country_code, country_name, confirmed, recovered, death FROM covid19 WHERE country_name = '%s' ORDER BY 1`, escapedCountryName),
	)

	if err == nil {
		for rows.Next() {
			var entry models.CountryEntry
			if rows.Scan(&entry.Timestamp, &entry.Code, &entry.Name, &entry.Confirmed, &entry.Recovered, &entry.Deaths) == nil {
				entries = append(entries, entry)
			}
		}
		_ = rows.Close()
	}
	return
}

// GetLatestForCountries gets the last entries for each specified country
func (store *PGCovidStore) GetLatestForCountries(countryNames []string) (entries map[string]models.CountryEntry, err error) {
	entries = make(map[string]models.CountryEntry)
	for _, countryName := range countryNames {
		const query = `SELECT time, country_name, country_code, confirmed, death, recovered FROM covid19 WHERE country_name = '%s' ORDER BY time DESC LIMIT 1`
		escapedCountryName := EscapeString(countryName)

		row := store.DB.Handle.QueryRow(fmt.Sprintf(query, escapedCountryName))

		var result = models.CountryEntry{}
		err = row.Scan(&result.Timestamp, &result.Name, &result.Code, &result.Confirmed, &result.Deaths, &result.Recovered)
		if err != nil && err != sql.ErrNoRows {
			log.WithError(err).Warning("failed to get entry from database")
			continue
		}

		entry, ok := entries[countryName]
		if ok == false {
			entry = models.CountryEntry{}
			entry.Timestamp = result.Timestamp
			entry.Name = result.Name
			entry.Code = result.Code
		}
		entry.Confirmed += result.Confirmed
		entry.Deaths += result.Deaths
		entry.Recovered += result.Recovered

		entries[countryName] = entry
	}
	return
}

// GetLatestForCountriesByTime gets the last entries for each specified country
func (store *PGCovidStore) GetLatestForCountriesByTime(countryNames []string, endTime time.Time) (entries map[string]models.CountryEntry, err error) {
	entries = make(map[string]models.CountryEntry)
	for _, countryName := range countryNames {
		const query = `SELECT time, country_name, country_code, confirmed, death, recovered FROM covid19 WHERE country_name = '%s' AND time <= '%s' ORDER BY time DESC LIMIT 1`
		escapedCountryName := EscapeString(countryName)

		row := store.DB.Handle.QueryRow(fmt.Sprintf(query, escapedCountryName, endTime.Format(time.RFC3339)))

		var result = models.CountryEntry{}
		err = row.Scan(&result.Timestamp, &result.Name, &result.Code, &result.Confirmed, &result.Deaths, &result.Recovered)
		if err != nil && err != sql.ErrNoRows {
			log.WithError(err).Warning("failed to get entry from database")
			continue
		}

		entry, ok := entries[countryName]
		if ok == false {
			entry = models.CountryEntry{}
			entry.Timestamp = result.Timestamp
			entry.Name = result.Name
			entry.Code = result.Code
		}
		entry.Confirmed += result.Confirmed
		entry.Deaths += result.Deaths
		entry.Recovered += result.Recovered

		entries[countryName] = entry
	}
	return
}

// Add inserts new entries in the database
func (store *PGCovidStore) Add(entries []models.CountryEntry) (err error) {
	var tx *sql.Tx
	tx, err = store.DB.Handle.Begin()

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
	return
}

// GetFirstEntry gets the timestamp of the first entry in the database
func (store *PGCovidStore) GetFirstEntry() (first time.Time, found bool, err error) {
	// TODO: SELECT MIN(time) may be more efficient, but makes Scan throw the following error:
	// "sql: Scan error on column index 0, name \"min\": unsupported Scan, storing driver.Value type <nil> into type *time.Time"
	const query = `SELECT time FROM covid19 ORDER BY 1 LIMIT 1`
	err = store.DB.Handle.QueryRow(query).Scan(&first)
	found = err == nil

	if err == sql.ErrNoRows {
		err = nil
	}
	return
}

// GetAllCountryNames gets all unique country names from the database
func (store *PGCovidStore) GetAllCountryNames() (names []string, err error) {
	return store.doLookup(`SELECT DISTINCT country_name FROM covid19 ORDER BY 1`)
}

func (store *PGCovidStore) doLookup(statement string) (names []string, err error) {
	var rows *sql.Rows
	rows, err = store.DB.Handle.Query(statement)

	if err == nil {
		for rows.Next() {
			var name string
			if rows.Scan(&name) == nil {
				names = append(names, name)
			}
		}
		_ = rows.Close()
	}

	return
}
