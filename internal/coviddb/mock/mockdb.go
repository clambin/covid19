package mock

import (
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	"covid19/internal/coviddb"
)

// DB mockapi database used for unit test tools
type DB struct {
	data []coviddb.CountryEntry
	lock sync.RWMutex
}

// Create a mockapi database
func Create(data []coviddb.CountryEntry) *DB {
	return &DB{data: data}
}

// List the records in the DB up to end date
func (dbh *DB) List(endDate time.Time) ([]coviddb.CountryEntry, error) {
	dbh.lock.RLock()
	defer dbh.lock.RUnlock()
	entries := make([]coviddb.CountryEntry, 0)

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
	dbh.lock.RLock()
	defer dbh.lock.RUnlock()
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
	dbh.lock.RLock()
	defer dbh.lock.RUnlock()
	first := time.Time{}
	for index, entry := range dbh.data {
		if index == 0 || entry.Timestamp.Before(first) {
			first = entry.Timestamp
		}
	}
	return first, nil
}

func (dbh *DB) GetLastBeforeDate(countryName string, before time.Time) (*coviddb.CountryEntry, error) {
	dbh.lock.RLock()
	defer dbh.lock.RUnlock()
	result := coviddb.CountryEntry{}
	for _, entry := range dbh.data {
		if entry.Name == countryName && entry.Timestamp.Before(before) && entry.Timestamp.After(result.Timestamp) {
			result.Timestamp = entry.Timestamp
			result.Code = entry.Code
			result.Name = entry.Name
			result.Confirmed = entry.Confirmed
			result.Deaths = entry.Deaths
			result.Recovered = entry.Recovered
		}
	}
	if result.Name == "" {
		return nil, nil
	}

	return &result, nil
}

// Add inserts all specified records in the covid19 database table
func (dbh *DB) Add(entries []coviddb.CountryEntry) error {
	dbh.lock.Lock()
	defer dbh.lock.Unlock()
	for _, entry := range entries {
		dbh.data = append(dbh.data, entry)
	}
	return nil
}
