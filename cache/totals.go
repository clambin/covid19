package cache

import (
	"github.com/clambin/covid19/models"
	"sort"
	"time"
)

// GetTotalCases calculates the total cases across all countries over time
func GetTotalCases(rows []*models.CountryEntry) (result []Entry) {
	var confirmed, recovered, deaths int64

	// Helper data structure to keep running count
	type covidData struct {
		Confirmed int64
		Recovered int64
		Deaths    int64
		Active    int64
	}

	// Group data by timestamp
	timeMap := make(map[time.Time][]*models.CountryEntry)
	for _, row := range rows {
		if timeMap[row.Timestamp] == nil {
			timeMap[row.Timestamp] = make([]*models.CountryEntry, 0, 300)
		}
		timeMap[row.Timestamp] = append(timeMap[row.Timestamp], row)
	}

	// Get all unique timestamps
	timestamps := make([]time.Time, 0, len(timeMap))
	for timestamp := range timeMap {
		timestamps = append(timestamps, timestamp)
	}
	sort.Slice(timestamps, func(i, j int) bool { return timestamps[i].Before(timestamps[j]) })

	// Go through each timestamp, record running total for each country & compute total cases
	countryMap := make(map[string]covidData)
	result = make([]Entry, 0, 365)
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
		result = append(result, Entry{
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
func GetTotalDeltas(entries []Entry) (result []Entry) {
	current := Entry{}
	for _, entry := range entries {
		result = append(result, Entry{
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
