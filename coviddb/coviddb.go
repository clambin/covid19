package coviddb

import (
	"database/sql"
	"fmt"
	"time"
	"github.com/mpvl/unique"

	_ "github.com/lib/pq"
)

type CovidDB struct {
	pgdb *sql.DB
}

func Connect(host string, port int, dbname string, user string, password string) (*CovidDB, error) {
	var coviddb CovidDB

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := sql.Open("postgres",  psqlInfo)

	if err == nil {
			coviddb.pgdb = db
	}

	return &coviddb, err
}

func (db *CovidDB) Close() {
	db.pgdb.Close()
}

type CovidData struct {
	Confirmed int64
	Recovered int64
	Deaths int64
}

type Entry struct {
	Timestamp time.Time
	Confirmed int64
	Recovered int64
	Deaths int64
}

type CountryEntry struct {
	Timestamp time.Time
	Code string
	Name string
	Confirmed int64
	Recovered int64
	Deaths int64
}

func (db *CovidDB) List() ([]CountryEntry, error) {
	entries := make([]CountryEntry, 0)

	rows, err := db.pgdb.Query("SELECT time, country_code, country_name, confirmed, recovered, death FROM covid19 ORDER BY 1")

	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var entry CountryEntry
			err = rows.Scan(&entry.Timestamp, &entry.Code, &entry.Name, &entry.Confirmed, &entry.Recovered, &entry.Deaths)
			if err != nil {
				panic(err)
			}
			entries = append(entries, entry)
		}
	}

	return entries, err
}

// Helper code to use unique.Sort()

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

func GetTotalCases (rows []CountryEntry) ([]Entry) {
	var confirmed, recovered, deaths int64

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
	countryMap := make(map[string]CovidData)
	consolidated := make([]Entry, 0)
	for _, timestamp := range timestamps {
		for _, row := range timeMap[timestamp] {
			countryMap[row.Code] = CovidData{Confirmed: row.Confirmed, Recovered: row.Recovered, Deaths: row.Deaths}
		}
		confirmed, recovered, deaths = 0, 0, 0
		for _, data := range countryMap {
			confirmed += data.Confirmed
			recovered += data.Recovered
			deaths    += data.Deaths
		}
		consolidated = append(consolidated, Entry{Timestamp: timestamp, Confirmed: confirmed, Recovered: recovered, Deaths: deaths})
	}

	return consolidated
}

func GetTotalDeltas (rows []Entry) ([]Entry) {
	deltas := make([]Entry, 0)

	var confirmed, recovered, deaths int64
	confirmed, recovered, deaths = 0, 0, 0
	for _, row := range rows {
		deltas = append(deltas, Entry{
			Timestamp: row.Timestamp,
			Confirmed: row.Confirmed - confirmed,
			Recovered: row.Recovered - recovered,
			Deaths:    row.Deaths    - deaths})
		confirmed, recovered, deaths = row.Confirmed,  row.Recovered, row.Deaths
	}

	return deltas
}
