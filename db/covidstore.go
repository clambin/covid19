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

// CovidStore represents a database holding COVID-19 statistics
//
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
	CountEntriesByTime(from, to time.Time) (entries []struct {
		Timestamp time.Time
		Count     int
	}, err error)
	GetTotalsPerDay() (entries []models.CountryEntry, err error)
}

var _ CovidStore = &PGCovidStore{}

// PGCovidStore implements CovidStore for Postgres databases
type PGCovidStore struct {
	DB *DB
}

// NewCovidStore creates a new PGCovidStore and initializes the database, if necessary
func NewCovidStore(db *DB) *PGCovidStore {
	return &PGCovidStore{DB: db}
}

const (
	queryStatement      = `SELECT time "timestamp", country_code "code", country_name "name", confirmed, recovered, death "deaths" FROM covid19 ORDER BY 1`
	queryWhereStatement = `SELECT time "timestamp", country_code "code", country_name "name", confirmed, recovered, death "deaths" FROM covid19 WHERE %s ORDER BY 1`
)

// GetAll returns all entries in the database, sorted by timestamp
func (store *PGCovidStore) GetAll() ([]models.CountryEntry, error) {
	var countryEntries []models.CountryEntry
	err := store.DB.Handle.Select(&countryEntries, queryStatement)
	if err != nil {
		err = fmt.Errorf("database query: %w", err)
	}
	return countryEntries, err
}

// GetAllForRange returns all entries in the database, sorted by timestamp
func (store *PGCovidStore) GetAllForRange(from, to time.Time) ([]models.CountryEntry, error) {
	var countryEntries []models.CountryEntry
	err := store.DB.Handle.Select(&countryEntries, fmt.Sprintf(queryWhereStatement, fmt.Sprintf(
		`time >= '%s' and time <= '%s'`,
		from.Format(time.RFC3339),
		to.Format(time.RFC3339),
	)))
	if err != nil {
		err = fmt.Errorf("database query: %w", err)
	}
	return countryEntries, err
}

// GetAllForCountryName returns all entries in the database, sorted by timestamp
func (store *PGCovidStore) GetAllForCountryName(countryName string) ([]models.CountryEntry, error) {
	var countryEntries []models.CountryEntry
	err := store.DB.Handle.Select(&countryEntries, fmt.Sprintf(queryWhereStatement, fmt.Sprintf(
		`country_name = '%s'`, escapeString(countryName))))
	if err != nil {
		err = fmt.Errorf("database query: %w", err)
	}
	return countryEntries, err
}

// GetLatestForCountries gets the last entries for each specified country
func (store *PGCovidStore) GetLatestForCountries(countryNames []string) (map[string]models.CountryEntry, error) {
	entries := make(map[string]models.CountryEntry)
	for _, countryName := range countryNames {
		var result []models.CountryEntry
		err := store.DB.Handle.Select(&result, fmt.Sprintf(queryWhereStatement+" DESC LIMIT 1",
			fmt.Sprintf(`country_name = '%s' `, escapeString(countryName))))
		if err != nil {
			return nil, fmt.Errorf("database query: %w", err)
		}
		if len(result) > 0 {
			entries[countryName] = result[0]
		}
	}
	return entries, nil
}

// GetLatestForCountriesByTime gets the last entries for each specified country
func (store *PGCovidStore) GetLatestForCountriesByTime(countryNames []string, endTime time.Time) (map[string]models.CountryEntry, error) {
	entries := make(map[string]models.CountryEntry)
	for _, countryName := range countryNames {
		var result []models.CountryEntry
		err := store.DB.Handle.Select(&result, fmt.Sprintf(queryWhereStatement, fmt.Sprintf(
			`country_name = '%s' AND time <= '%s'`, escapeString(countryName), endTime.Format(time.RFC3339)))+" DESC LIMIT 1")
		if err != nil {
			return nil, fmt.Errorf("database query: %w", err)
		}
		if len(result) > 0 {
			entries[countryName] = result[0]
		}
	}
	return entries, nil
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

// GetFirstEntry gets the timestamp of the first entry in the database
func (store *PGCovidStore) GetFirstEntry() (time.Time, bool, error) {
	// TODO: SELECT MIN(time) may be more efficient, but makes Scan throw the following error:
	// "sql: Scan error on column index 0, name \"min\": unsupported Scan, storing driver.Value type <nil> into type *time.Time"
	var first time.Time
	err := store.DB.Handle.Get(&first, `SELECT time FROM covid19 ORDER BY 1 LIMIT 1`)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = nil
		}
		return time.Time{}, false, err
	}
	return first, true, nil
}

// GetAllCountryNames gets all unique country names from the database
func (store *PGCovidStore) GetAllCountryNames() (names []string, err error) {
	err = store.DB.Handle.Select(&names, `SELECT DISTINCT country_name FROM covid19 ORDER BY 1`)
	return names, err
}

// CountEntriesByTime counts updates per timestamp
func (store *PGCovidStore) CountEntriesByTime(from, to time.Time) ([]struct {
	Timestamp time.Time
	Count     int
}, error) {
	var updates []struct {
		Timestamp time.Time
		Count     int
	}

	err := store.DB.Handle.Select(&updates, fmt.Sprintf(`SELECT time "timestamp", COUNT(*) "count" FROM covid19 %s GROUP BY time ORDER BY time`, makeTimestampClause(from, to)))
	return updates, err
}

// GetTotalsPerDay returns the total new cases per day across all countries
func (store *PGCovidStore) GetTotalsPerDay() ([]models.CountryEntry, error) {
	var entries []models.CountryEntry
	err := store.DB.Handle.Select(&entries, `SELECT time "timestamp", SUM(confirmed) "confirmed", SUM(death) "deaths" FROM covid19 GROUP BY time ORDER BY time`)
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
		clause = "WHERE " + strings.Join(conditions, " AND ")
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
