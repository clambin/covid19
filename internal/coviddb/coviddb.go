package coviddb

import (
	"fmt"
	"time"
	"database/sql"

	"github.com/lib/pq"
	"github.com/mpvl/unique"
	log "github.com/sirupsen/logrus"

)

// CovidDB interface representing a Covid batabase
type CovidDB interface {
	List(time.Time) ([]CountryEntry, error)
	ListLatestByCountry() (map[string]time.Time, error)
	Add([]CountryEntry) (error)
}

// Indexes for the output arrays of GetTotalCases / GetTotalDeltas
const (
	CONFIRMED = 0
	RECOVERED = 1
	DEATHS    = 2
	ACTIVE    = 3

	VALUE     = 0
	TIMESTAMP = 1
)

// PostgresCovidDB implementation of CovidDB
type PostgresCovidDB struct {
	psqlInfo    string
	initialized bool
}

// Create a PostgresCovidDB object
func Create(host string, port int, database string, user string, password string) (*PostgresCovidDB) {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, database)

	return &PostgresCovidDB{psqlInfo: psqlInfo, initialized: false}
}

// CountryEntry represents one row in the Covid DB
type CountryEntry struct {
	Timestamp time.Time
	Code string
	Name string
	Confirmed int64
	Recovered int64
	Deaths int64
}

// List retrieved all records from the database up to endDate
func (db *PostgresCovidDB) List(endDate time.Time) ([]CountryEntry, error) {
	entries := make([]CountryEntry, 0)

	if err := db.initializeDB(); err != nil {
		return entries, err
	}

	dbh, err := sql.Open("postgres", db.psqlInfo)

	if err == nil {
		defer dbh.Close()

		rows, err := dbh.Query(fmt.Sprintf(
			"SELECT time, country_code, country_name, confirmed, recovered, death FROM covid19 WHERE time <= '%s' ORDER BY 1",
			endDate.Format("2006-01-02 15:04:05")))

		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var entry CountryEntry
				err = rows.Scan(&entry.Timestamp, &entry.Code, &entry.Name, &entry.Confirmed, &entry.Recovered, &entry.Deaths)
				if err != nil { break }
				entries = append(entries, entry)
			}
			log.Debugf("Found %d records", len(entries))
		}
	}

	return entries, err
}

// ListLatestByCountry returns the timestamp of each country's last update
func (db *PostgresCovidDB) ListLatestByCountry() (map[string]time.Time, error) {
	entries := make(map[string]time.Time, 0)

	if err := db.initializeDB(); err != nil {
		return entries, err
	}

	dbh, err := sql.Open("postgres", db.psqlInfo)

	if err == nil {
		defer dbh.Close()

		rows, err := dbh.Query(fmt.Sprintf("SELECT country_name, MAX(time) FROM covid19 GROUP BY country_name"))

		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var (
					country string
					timestamp time.Time
				)
				err = rows.Scan(&country, &timestamp)
				if err != nil { break }
				entries[country] = timestamp
			}
			log.Debugf("Found %d records", len(entries))
		}
	}

	return entries, err
}

// Add inserts all specified records in the covid19 database table
func (db *PostgresCovidDB) Add(entries []CountryEntry) (error) {
	if err := db.initializeDB(); err != nil {
		return err
	}

	dbh, err := sql.Open("postgres", db.psqlInfo)

	if err != nil { return err }
	defer dbh.Close()

	txn, err := dbh.Begin()
	if err != nil { return err }

	stmt, err := txn.Prepare(pq.CopyIn("covid19", "time", "country_code", "country_name", "confirmed", "death", "recovered"))
	if err != nil { return err }

	for _, entry := range entries {
		_, err = stmt.Exec(entry.Timestamp, entry.Code, entry.Name, entry.Confirmed, entry.Deaths, entry.Recovered)
		if err != nil { return err }
	}

	_, err = stmt.Exec()
	if err != nil { return err }

	err = stmt.Close()
	if err != nil { return err }

	err = txn.Commit()
	if err != nil { return err }

	return nil
}

// Helper code for unique.Sort()
type timestampSlice struct { P *[]time.Time }

func (p timestampSlice) Len() int {
	return len(*p.P)
}

func (p timestampSlice) Less(i, j int) bool {
	return (*p.P)[i].Before((*p.P)[j])
}

func (p timestampSlice) Swap(i, j int) {
	(*p.P)[i], (*p.P)[j] = (*p.P)[j], (*p.P)[i]
}

func (p timestampSlice) Truncate(n int) {
		(*p.P) = (*p.P)[:n]
}

// GetTotalCases calculates the total cases cross all countries over time
// Output is structured for easy export to HTTP Response (JSON)
func GetTotalCases (rows []CountryEntry) ([][][]int64) {
	var confirmed, recovered, deaths int64

	// Helper datastructure to keep running count
	type covidData struct {
		Confirmed int64
		Recovered int64
		Deaths int64
		Active int64
	}

	// Group data by timestamp
	timeMap := make(map[time.Time][]CountryEntry)
	timestamps := make([]time.Time, 0)
	for _, row := range rows {
		if timeMap[row.Timestamp] == nil {
			timeMap[row.Timestamp] = make([]CountryEntry, 0)
		}
		timeMap[row.Timestamp] = append(timeMap[row.Timestamp], row)
		timestamps = append(timestamps, row.Timestamp)
	}
	unique.Sort(timestampSlice{&timestamps})

	// Go through each timestamp, record running total for each country & compute total cases
	countryMap := make(map[string]covidData)
	consolidated := make([][][]int64, 4)
	for _, timestamp := range timestamps {
		for _, row := range timeMap[timestamp] {
			countryMap[row.Code] = covidData{Confirmed: row.Confirmed, Recovered: row.Recovered, Deaths: row.Deaths}
		}
		confirmed, recovered, deaths = 0, 0, 0
		for _, data := range countryMap {
			confirmed += data.Confirmed
			recovered += data.Recovered
			deaths    += data.Deaths
		}
		epoch := timestamp.UnixNano() / 1000000
		consolidated[CONFIRMED] = append(consolidated[CONFIRMED], []int64{confirmed,                      epoch})
		consolidated[RECOVERED] = append(consolidated[RECOVERED], []int64{recovered,                      epoch})
		consolidated[DEATHS]    = append(consolidated[DEATHS],    []int64{deaths,                         epoch})
		consolidated[ACTIVE]    = append(consolidated[ACTIVE],    []int64{confirmed - recovered - deaths, epoch})
	}

	return consolidated
}

// GetTotalDeltas calculates deltas of cases returned by GetTotalCases
// Output is structured for easy export to HTTP Response (JSON)
func GetTotalDeltas (rows [][]int64) ([][]int64) {
	deltas := make([][]int64, 0)

	var value int64
	value = 0
	for _, row := range rows {
		deltas = append(deltas, []int64{row[0] - value, row[1]})
		value = row[0]
	}

	return deltas
}

// initializeDB created the required tables
func (db *PostgresCovidDB) initializeDB() (error) {
	if db.initialized {
		return nil
	}

	dbh, err := sql.Open("postgres", db.psqlInfo)

	if err != nil { return err }
	defer dbh.Close()

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

	if err == nil {
		db.initialized = true
	}

	return err
}
