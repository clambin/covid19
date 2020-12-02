package mock

import (
	"time"

	log     "github.com/sirupsen/logrus"

	"covid19api/pkg/coviddb"
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

