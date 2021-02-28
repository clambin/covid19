package covidcache

import (
	"covid19/internal/coviddb"
	"github.com/mpvl/unique"
	"time"
)

// GetTotalCases calculates the total cases across all countries over time
func GetTotalCases(rows []coviddb.CountryEntry) (result []CacheEntry) {
	var confirmed, recovered, deaths int64

	// Helper data structure to keep running count
	type covidData struct {
		Confirmed int64
		Recovered int64
		Deaths    int64
		Active    int64
	}

	// Group data by timestamp
	timeMap := make(map[time.Time][]coviddb.CountryEntry)
	timestamps := make([]time.Time, 0)
	for _, row := range rows {
		if timeMap[row.Timestamp] == nil {
			timeMap[row.Timestamp] = make([]coviddb.CountryEntry, 0)
		}
		timeMap[row.Timestamp] = append(timeMap[row.Timestamp], row)
		timestamps = append(timestamps, row.Timestamp)
	}
	unique.Sort(timestampSlice{&timestamps})

	// Go through each timestamp, record running total for each country & compute total cases
	countryMap := make(map[string]covidData)
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
		result = append(result, CacheEntry{
			Timestamp: timestamp,
			Confirmed: confirmed,
			Recovered: recovered,
			Deaths:    deaths,
			Active:    confirmed - recovered - deaths,
		})
	}
	return
}

// GetTotalDeltas calculates deltas of cases returned by GetTotalCases
func GetTotalDeltas(entries []CacheEntry) (result []CacheEntry) {
	current := CacheEntry{}
	for _, entry := range entries {
		result = append(result, CacheEntry{
			Timestamp: entry.Timestamp,
			Confirmed: entry.Confirmed - current.Confirmed,
			Recovered: entry.Recovered - current.Recovered,
			Deaths:    entry.Deaths - current.Deaths,
			Active:    entry.Active - current.Active,
		})
		current = entry
	}
	return
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
