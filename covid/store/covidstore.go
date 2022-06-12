package store

import (
	"database/sql"
	"errors"
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
	GetAllForRange(from, to time.Time) (entries []models.CountryEntry, err error)
	GetAllForCountryName(name string) (entries []models.CountryEntry, err error)
	GetLatestForCountries(countryNames []string) (entries map[string]models.CountryEntry, err error)
	GetLatestForCountriesByTime(countryNames []string, endTime time.Time) (entries map[string]models.CountryEntry, err error)
	GetFirstEntry() (first time.Time, found bool, err error)
	Add(entries []models.CountryEntry) (err error)
	GetAllCountryNames() (names []string, err error)
	CountEntriesByTime(from, to time.Time) (count map[time.Time]int, err error)
	GetTotalsPerDay() (entries []models.CountryEntry, err error)
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
	return store.queryEntries(`ORDER BY 1`)
}

// GetAllForRange returns all entries in the database, sorted by timestamp
func (store *PGCovidStore) GetAllForRange(from, to time.Time) (entries []models.CountryEntry, err error) {
	return store.queryEntries(
		fmt.Sprintf(
			`WHERE time >= '%s' and time <= '%s' ORDER BY 1`,
			from.Format(time.RFC3339),
			to.Format(time.RFC3339),
		),
	)
}

// GetAllForCountryName returns all entries in the database, sorted by timestamp
func (store *PGCovidStore) GetAllForCountryName(countryName string) (entries []models.CountryEntry, err error) {
	return store.queryEntries(
		fmt.Sprintf(
			`WHERE country_name = '%s' ORDER BY 1`,
			EscapeString(countryName),
		),
	)
}

// GetLatestForCountries gets the last entries for each specified country
func (store *PGCovidStore) GetLatestForCountries(countryNames []string) (entries map[string]models.CountryEntry, err error) {
	entries = make(map[string]models.CountryEntry)
	for _, countryName := range countryNames {
		var result []models.CountryEntry
		result, err = store.queryEntries(fmt.Sprintf(
			`WHERE country_name = '%s' ORDER BY time DESC LIMIT 1`,
			EscapeString(countryName),
		))

		if err == nil && len(result) > 0 {
			entries[countryName] = result[0]
			err = nil
		}
	}
	return
}

// GetLatestForCountriesByTime gets the last entries for each specified country
func (store *PGCovidStore) GetLatestForCountriesByTime(countryNames []string, endTime time.Time) (entries map[string]models.CountryEntry, err error) {
	entries = make(map[string]models.CountryEntry)
	for _, countryName := range countryNames {
		var result []models.CountryEntry
		result, err = store.queryEntries(fmt.Sprintf(
			`WHERE country_name = '%s' AND time <= '%s' ORDER BY time DESC LIMIT 1`,
			EscapeString(countryName),
			endTime.Format(time.RFC3339),
		))

		if err == nil && len(result) > 0 {
			entries[countryName] = result[0]
		}
	}
	return
}

func (store *PGCovidStore) queryEntries(conditions string) (entries []models.CountryEntry, err error) {
	if len(conditions) > 0 {
		conditions = " " + conditions
	}
	var rows *sql.Rows
	rows, err = store.DB.Handle.Query(`SELECT time, country_name, country_code, confirmed, recovered, death FROM covid19` + conditions)

	if err == nil {
		for rows.Next() {
			var entry models.CountryEntry
			if rows.Scan(&entry.Timestamp, &entry.Code, &entry.Name, &entry.Confirmed, &entry.Recovered, &entry.Deaths) == nil {
				entries = append(entries, entry)
			}
		}
		_ = rows.Close()
	}

	if err == sql.ErrNoRows {
		err = nil
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

// CountEntriesByTime counts updates per timestamp
func (store *PGCovidStore) CountEntriesByTime(from, to time.Time) (updates map[time.Time]int, err error) {
	var rows *sql.Rows
	rows, err = store.DB.Handle.Query(fmt.Sprintf(`SELECT time, COUNT(*) FROM covid19 %s GROUP BY time`, makeTimestampClause(from, to)))

	if errors.Is(err, sql.ErrNoRows) {
		return updates, nil
	}
	if err != nil {
		return
	}

	updates = make(map[time.Time]int)
	for rows.Next() {
		var timestamp time.Time
		var count int

		if rows.Scan(&timestamp, &count) == nil {
			updates[timestamp] = count
		}
	}
	_ = rows.Close()

	return
}

// GetTotalsPerDay returns the total new cases per day across all countries
func (store *PGCovidStore) GetTotalsPerDay() (entries []models.CountryEntry, err error) {
	var rows *sql.Rows
	rows, err = store.DB.Handle.Query(`SELECT time, SUM(confirmed), SUM(death) FROM covid19 GROUP BY time ORDER BY time`)

	if errors.Is(err, sql.ErrNoRows) {
		return entries, nil
	}
	if err != nil {
		return
	}

	for rows.Next() {
		var entry models.CountryEntry
		if rows.Scan(&entry.Timestamp, &entry.Confirmed, &entry.Deaths) == nil {
			entries = append(entries, entry)
		}
	}
	_ = rows.Close()

	return
}
