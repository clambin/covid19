package db

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/clambin/covid19/models"
	"github.com/lib/pq"
	"strings"
	"time"
)

// PGCovidStore implements CovidStore for Postgres databases
type PGCovidStore struct {
	DB *DB
}

// NewCovidStore creates a new PGCovidStore and initializes the database, if necessary
func NewCovidStore(db *DB) *PGCovidStore {
	return &PGCovidStore{DB: db}
}

const (
	queryStatement = `SELECT time "timestamp", country_code "code", country_name "name", confirmed, recovered, death "deaths" FROM covid19`
)

// GetAllForRange returns all entries in the database, sorted by timestamp
func (store *PGCovidStore) GetAllForRange(from, to time.Time) ([]models.CountryEntry, error) {
	var countryEntries []models.CountryEntry
	err := store.DB.Handle.Select(&countryEntries, queryStatement+` WHERE `+makeTimestampClause(from, to)+` ORDER BY 1`)
	return countryEntries, err
}

// GetAllForCountryName returns all entries in the database, sorted by timestamp
func (store *PGCovidStore) GetAllForCountryName(countryName string) ([]models.CountryEntry, error) {
	var countryEntries []models.CountryEntry
	err := store.DB.Handle.Select(&countryEntries, queryStatement+` WHERE country_name = '`+escapeString(countryName)+`' ORDER BY 1`)
	return countryEntries, err
}

// GetLatestForCountries gets the last entries for each country up the specified endTime.
// If endTime is time.Time{}, it will get the latest entries up to the current time.
func (store *PGCovidStore) GetLatestForCountries(endTime time.Time) (map[string]models.CountryEntry, error) {
	countryNames, err := store.GetAllCountryNames()
	if err != nil {
		return nil, err
	}

	if endTime.IsZero() {
		endTime = time.Now()
	}

	entries := make(map[string]models.CountryEntry)
	for _, countryName := range countryNames {
		var entry models.CountryEntry
		entry, err = store.getLatestForCountry(countryName, endTime)
		if errors.Is(err, sql.ErrNoRows) {
			continue
		}
		if err != nil {
			break
		}
		entries[countryName] = entry
	}
	return entries, err
}

func (store *PGCovidStore) getLatestForCountry(countryName string, endTime time.Time) (models.CountryEntry, error) {
	timestampClause := makeTimestampClause(time.Time{}, endTime)
	if timestampClause != "" {
		timestampClause = " AND " + timestampClause
	}
	statement := queryStatement + ` WHERE country_name = '%s'` + timestampClause + ` ORDER BY 1 DESC`

	var entry models.CountryEntry
	err := store.DB.Handle.Get(&entry, fmt.Sprintf(statement, escapeString(countryName)))
	return entry, err
}

// Add inserts new entries in the database
func (store *PGCovidStore) Add(entries []models.CountryEntry) error {
	tx := store.DB.Handle.MustBegin()
	defer func() {
		// will be ignored if we commit before the function returns
		_ = tx.Rollback()
	}()

	stmt, err := tx.Prepare(pq.CopyIn("covid19", "time", "country_code", "country_name", "confirmed", "death", "recovered"))
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if _, err = stmt.Exec(entry.Timestamp, entry.Code, entry.Name, entry.Confirmed, entry.Deaths, entry.Recovered); err != nil {
			return err
		}
	}

	if _, err = stmt.Exec(); err == nil {
		err = tx.Commit()
	}
	return err
}

// Rows returns the number of rows in the store
func (store *PGCovidStore) Rows() (int, error) {
	var rows int
	err := store.DB.Handle.Get(&rows, `SELECT COUNT(*) AS rows FROM covid19`)
	return rows, err
}

// GetAllCountryNames gets all unique country names from the database
func (store *PGCovidStore) GetAllCountryNames() (names []string, err error) {
	err = store.DB.Handle.Select(&names, `SELECT DISTINCT country_name FROM covid19 ORDER BY 1`)
	return names, err
}

type TimestampCount struct {
	Timestamp time.Time
	Count     int
}

// CountEntriesByTime counts updates per timestamp
func (store *PGCovidStore) CountEntriesByTime(from, to time.Time) ([]TimestampCount, error) {
	var updates []TimestampCount
	whereClause := makeTimestampClause(from, to)
	if whereClause != "" {
		whereClause = " WHERE " + whereClause
	}

	err := store.DB.Handle.Select(&updates, `SELECT time AS "timestamp", COUNT(*) "count" FROM covid19 `+whereClause+` GROUP BY time ORDER BY time`)
	return updates, err
}

// GetTotalsPerDay returns the total new cases per day across all countries
func (store *PGCovidStore) GetTotalsPerDay() ([]models.CountryEntry, error) {
	var entries []models.CountryEntry
	err := store.DB.Handle.Select(&entries, `SELECT time AS "timestamp", SUM(confirmed) AS "confirmed", SUM(death) AS "deaths" FROM covid19 GROUP BY time ORDER BY time`)
	return entries, err
}

func makeTimestampClause(from, to time.Time) (clause string) {
	var conditions []string
	if !from.IsZero() {
		conditions = append(conditions, fmt.Sprintf("time >= '%s'", from.Format(time.RFC3339)))
	}
	if !to.IsZero() {
		conditions = append(conditions, fmt.Sprintf("time <= '%s'", to.Format(time.RFC3339)))
	}
	if len(conditions) > 0 {
		clause = strings.Join(conditions, " AND ")
	}
	return
}

func escapeString(input string) (output string) {
	for _, c := range input {
		if c == '\'' {
			output += "'"
		}
		output += string(c)
	}
	return
}
