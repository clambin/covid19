package covidhandler

import (
	"time"

	"github.com/mpvl/unique"

	"covid19/internal/covid/db"
)

// Indexes for the output arrays of GetTotalCases / GetTotalDeltas
const (
	CONFIRMED = 0
	RECOVERED = 1
	DEATHS    = 2
	ACTIVE    = 3

	VALUE     = 0
	TIMESTAMP = 1
)

// GetTotalCases calculates the total cases cross all countries over time
// Output is structured for easy export to HTTP Response (JSON)
func GetTotalCases(rows []db.CountryEntry) [][][2]int64 {
	var confirmed, recovered, deaths int64

	// Helper datastructure to keep running count
	type covidData struct {
		Confirmed int64
		Recovered int64
		Deaths    int64
		Active    int64
	}

	// Group data by timestamp
	timeMap := make(map[time.Time][]db.CountryEntry)
	timestamps := make([]time.Time, 0)
	for _, row := range rows {
		if timeMap[row.Timestamp] == nil {
			timeMap[row.Timestamp] = make([]db.CountryEntry, 0)
		}
		timeMap[row.Timestamp] = append(timeMap[row.Timestamp], row)
		timestamps = append(timestamps, row.Timestamp)
	}
	unique.Sort(timestampSlice{&timestamps})

	// Go through each timestamp, record running total for each country & compute total cases
	countryMap := make(map[string]covidData)
	consolidated := make([][][2]int64, 4)
	for _, timestamp := range timestamps {
		for _, row := range timeMap[timestamp] {
			countryMap[row.Code] = covidData{Confirmed: row.Confirmed, Recovered: row.Recovered, Deaths: row.Deaths}
		}
		confirmed, recovered, deaths = 0, 0, 0
		for _, data := range countryMap {
			confirmed += data.Confirmed
			recovered += data.Recovered
			deaths += data.Deaths
		}
		epoch := timestamp.UnixNano() / 1000000
		consolidated[CONFIRMED] = append(consolidated[CONFIRMED], [2]int64{confirmed, epoch})
		consolidated[RECOVERED] = append(consolidated[RECOVERED], [2]int64{recovered, epoch})
		consolidated[DEATHS] = append(consolidated[DEATHS], [2]int64{deaths, epoch})
		consolidated[ACTIVE] = append(consolidated[ACTIVE], [2]int64{confirmed - recovered - deaths, epoch})
	}

	return consolidated
}

// GetTotalDeltas calculates deltas of cases returned by GetTotalCases
// Output is structured for easy export to HTTP Response (JSON)
func GetTotalDeltas(rows [][2]int64) [][2]int64 {
	deltas := make([][2]int64, 0)

	var value int64
	value = 0
	for _, row := range rows {
		deltas = append(deltas, [2]int64{row[0] - value, row[1]})
		value = row[0]
	}

	return deltas
}

// Helper code for unique.Sort()
type timestampSlice struct{ P *[]time.Time }

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
	*p.P = (*p.P)[:n]
}
