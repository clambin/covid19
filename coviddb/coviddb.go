package coviddb

import (
	"database/sql"
	"fmt"
	"time"
	"github.com/mpvl/unique"

	_ "github.com/lib/pq"
)

const (
	CONFIRMED = 0
	RECOVERED = 1
	DEATHS    = 2
	ACTIVE    = 3

	TIMESTAMP = 0
	VALUE     = 1
)

type CovidDB struct {
	psqlInfo  string
}

func Create(host string, port int, database string, user string, password string) (*CovidDB) {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, database)

	return &CovidDB{psqlInfo: psqlInfo}
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

	dbh, err := sql.Open("postgres", db.psqlInfo)

	if err == nil {
		rows, err := dbh.Query("SELECT time, country_code, country_name, confirmed, recovered, death FROM covid19 ORDER BY 1")

		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var entry CountryEntry
				err = rows.Scan(&entry.Timestamp, &entry.Code, &entry.Name, &entry.Confirmed, &entry.Recovered, &entry.Deaths)
				if err != nil {
					break
				} else {
					entries = append(entries, entry)
				}
			}
		}
		dbh.Close()
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

type covidData struct {
	Confirmed int64
	Recovered int64
	Deaths int64
	Active int64
}

type Entry struct {
	Timestamp time.Time
	Value int64
}

func GetTotalCases (rows []CountryEntry) ([][][]int64) {
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
		// TODO: convert timestamp to grafana representatation (msec since epoch?)
		epoch := timestamp.UnixNano() / 1000000
		consolidated[CONFIRMED] = append(consolidated[CONFIRMED], []int64{epoch, confirmed})
		consolidated[RECOVERED] = append(consolidated[RECOVERED], []int64{epoch, recovered})
		consolidated[DEATHS]    = append(consolidated[DEATHS],    []int64{epoch, deaths})
		consolidated[ACTIVE]    = append(consolidated[ACTIVE],    []int64{epoch, confirmed - recovered - deaths})
	}

	return consolidated
}

func GetTotalDeltas (rows [][]int64) ([][]int64) {
	deltas := make([][]int64, 0)

	var value int64
	value = 0
	for _, row := range rows {
		deltas = append(deltas, []int64{row[0], row[1] - value})
		value = row[1]
	}

	return deltas
}

