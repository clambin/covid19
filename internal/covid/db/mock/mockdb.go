package mock

import (
	"time"

	log "github.com/sirupsen/logrus"

	"covid19/internal/covid/db"
)

// DB mock database used for unit test tools
type DB struct {
	data []db.CountryEntry
}

// Create a mock database
func Create(data []db.CountryEntry) *DB {
	return &DB{data: data}
}

// List the records in the DB up to end date
func (dbh *DB) List(endDate time.Time) ([]db.CountryEntry, error) {
	entries := make([]db.CountryEntry, 0)

	for _, entry := range dbh.data {
		if entry.Timestamp.Before(endDate) {
			entries = append(entries, entry)
		} else {
			log.Debugf("Dropping '%s'", entry.Timestamp)
		}
	}

	return entries, nil
}

// ListLatestByCountry lists the last date per country
func (dbh *DB) ListLatestByCountry() (map[string]time.Time, error) {
	entries := make(map[string]time.Time, 0)

	for _, entry := range dbh.data {
		record, ok := entries[entry.Name]
		if ok == false || record.Before(entry.Timestamp) {
			entries[entry.Name] = entry.Timestamp
		}
	}

	return entries, nil
}

// GetFirstEntry returns the timestamp of the first entry
func (dbh *DB) GetFirstEntry() (time.Time, error) {
	first := time.Time{}
	for index, entry := range dbh.data {
		if index == 0 || entry.Timestamp.Before(first) {
			first = entry.Timestamp
		}
	}
	return first, nil
}

// Add inserts all specified records in the covid19 database table
func (dbh *DB) Add(entries []db.CountryEntry) error {
	for _, entry := range entries {
		dbh.data = append(dbh.data, entry)
	}
	return nil
}
