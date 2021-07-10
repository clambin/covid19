package mock

import (
	"github.com/clambin/covid19/internal/coviddb"
	log "github.com/sirupsen/logrus"
	"sync"
	"time"
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
func (dbh *DB) GetFirstEntry() (first time.Time, found bool, err error) {
	dbh.lock.RLock()
	defer dbh.lock.RUnlock()
	first = time.Time{}
	for index, entry := range dbh.data {
		if index == 0 || entry.Timestamp.Before(first) {
			found = true
			first = entry.Timestamp
		}
	}
	return
}

func (dbh *DB) GetLastBeforeDate(countryName string, before time.Time) (result *coviddb.CountryEntry, found bool, err error) {
	dbh.lock.RLock()
	defer dbh.lock.RUnlock()
	result = &coviddb.CountryEntry{}
	for _, entry := range dbh.data {
		if entry.Name == countryName && entry.Timestamp.Before(before) && entry.Timestamp.After(result.Timestamp) {
			result.Timestamp = entry.Timestamp
			result.Code = entry.Code
			result.Name = entry.Name
			result.Confirmed = entry.Confirmed
			result.Deaths = entry.Deaths
			result.Recovered = entry.Recovered
			found = true
		}
	}
	return
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

// GetAllCountryCodes returns all country codes that have an entry in the covid19 database table
func (dbh *DB) GetAllCountryCodes() (codes []string, err error) {
	dbh.lock.Lock()
	defer dbh.lock.Unlock()
	added := make(map[string]struct{})
	for _, entry := range dbh.data {
		if _, found := added[entry.Code]; found == false {
			codes = append(codes, entry.Code)
			added[entry.Code] = struct{}{}
		}
	}

	return
}
