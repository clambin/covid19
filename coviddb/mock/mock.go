package mock

import (
	"time"

	log     "github.com/sirupsen/logrus"

	"covid19api/coviddb"
)

type MockCovidDB struct {
	data []coviddb.CountryEntry
}

func Create(data []coviddb.CountryEntry) (MockCovidDB) {
	return MockCovidDB{data: data}
}

func (db MockCovidDB) List(enddate time.Time) ([]coviddb.CountryEntry, error) {
	entries := make([]coviddb.CountryEntry, 0)

	for _, entry := range db.data {
		if entry.Timestamp.Before(enddate) {
			entries = append(entries, entry)
		} else {
			log.Debugf("Dropping '%s'", entry.Timestamp)
		}
	}

	return entries, nil
}

