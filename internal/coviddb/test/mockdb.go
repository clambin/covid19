package test

import (
	"time"

	log "github.com/sirupsen/logrus"

	"covid19/internal/coviddb"
)

// DB mock database used for unit test tools
type DB struct {
	data []coviddb.CountryEntry
}

// Create a mock DB
func Create(data []coviddb.CountryEntry) *DB {
	return &DB{data: data}
}

func CreateWithData() *DB {
	return &DB{data: []coviddb.CountryEntry{
		{
			Timestamp: parseDate("2020-11-01T00:00:00.000Z"),
			Code:      "BE",
			Name:      "Belgium",
			Confirmed: 1,
			Recovered: 0,
			Deaths:    0},
		{
			Timestamp: parseDate("2020-11-02T00:00:00.000Z"),
			Code:      "US",
			Name:      "United States",
			Confirmed: 3,
			Recovered: 0,
			Deaths:    0},
		{
			Timestamp: parseDate("2020-11-02T00:00:00.000Z"),
			Code:      "BE",
			Name:      "Belgium",
			Confirmed: 3,
			Recovered: 1,
			Deaths:    0},
		{
			Timestamp: parseDate("2020-11-04T00:00:00.000Z"),
			Code:      "US",
			Name:      "United States",
			Confirmed: 10,
			Recovered: 5,
			Deaths:    1},
	}}
}

func parseDate(dateString string) time.Time {
	date, _ := time.Parse("2006-01-02T15:04:05.000Z", dateString)
	return date
}

// List the records in the DB up to end date
func (db *DB) List(endDate time.Time) ([]coviddb.CountryEntry, error) {
	entries := make([]coviddb.CountryEntry, 0)

	for _, entry := range db.data {
		if entry.Timestamp.Before(endDate) {
			entries = append(entries, entry)
		} else {
			log.Debugf("Dropping '%s'", entry.Timestamp)
		}
	}

	return entries, nil
}

// ListLatestByCountry lists the last date per country
func (db *DB) ListLatestByCountry() (map[string]time.Time, error) {
	entries := make(map[string]time.Time, 0)

	for _, entry := range db.data {
		record, ok := entries[entry.Name]
		if ok == false || record.Before(entry.Timestamp) {
			entries[entry.Name] = entry.Timestamp
		}
	}

	return entries, nil
}

// GetFirstEntry returns the timestamp of the first entry
func (db *DB) GetFirstEntry() (time.Time, error) {
	first := time.Time{}
	for index, entry := range db.data {
		if index == 0 || entry.Timestamp.Before(first) {
			first = entry.Timestamp
		}
	}
	return first, nil
}

// Add inserts all specified records in the covid19 database table
func (db *DB) Add(entries []coviddb.CountryEntry) error {
	for _, entry := range entries {
		db.data = append(db.data, entry)
	}
	return nil
}
