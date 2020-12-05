package mock

import (
	"time"

	log     "github.com/sirupsen/logrus"

	"covid19/internal/coviddb"
)

// CovidDB mock database used for unittesting 
type CovidDB struct {
	data []coviddb.CountryEntry
}

// Create a mock CovidDB
func Create(data []coviddb.CountryEntry) (*CovidDB) {
	return &CovidDB{data: data}
}

// List the records in the DB up to enddate
func (db *CovidDB) List(endDate time.Time) ([]coviddb.CountryEntry, error) {
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
func (db *CovidDB) ListLatestByCountry()  (map[string]time.Time, error) {
	entries := make(map[string]time.Time, 0)

	for _, entry := range db.data {
		record, ok := entries[entry.Name]
		if ok == false || record.Before(entry.Timestamp) {
			entries[entry.Name]  = entry.Timestamp
		}
	}

	return entries, nil
}

// Add inserts all specified records in the covid19 database table
func (db *CovidDB) Add(entries []coviddb.CountryEntry)  (error) {
	for _, entry := range entries {
		db.data = append(db.data, entry)
	}
	return nil
}
