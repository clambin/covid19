package mock

import (
	"time"

	log     "github.com/sirupsen/logrus"

	"covid19/internal/covid"
)

// DB mock database used for unittesting 
type DB struct {
	data []covid.CountryEntry
}

// Create a mock DB
func Create(data []covid.CountryEntry) (*DB) {
	return &DB{data: data}
}

// List the records in the DB up to enddate
func (db *DB) List(endDate time.Time) ([]covid.CountryEntry, error) {
	entries := make([]covid.CountryEntry, 0)

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
func (db *DB) ListLatestByCountry()  (map[string]time.Time, error) {
	entries := make(map[string]time.Time, 0)

	for _, entry := range db.data {
		record, ok := entries[entry.Name]
		if ok == false || record.Before(entry.Timestamp) {
			entries[entry.Name]  = entry.Timestamp
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
func (db *DB) Add(entries []covid.CountryEntry)  (error) {
	for _, entry := range entries {
		db.data = append(db.data, entry)
	}
	return nil
}
