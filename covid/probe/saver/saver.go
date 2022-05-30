package saver

import (
	"fmt"
	"github.com/clambin/covid19/covid/store"
	"github.com/clambin/covid19/models"
	log "github.com/sirupsen/logrus"
	"sort"
)

// Saver stores new entries in the database
//go:generate mockery --name Saver
type Saver interface {
	SaveNewEntries(entries []models.CountryEntry) (newEntries []models.CountryEntry, err error)
}

// StoreSaver implements the Saver interface for CovidStore
type StoreSaver struct {
	Store store.CovidStore
}

var _ Saver = &StoreSaver{}

// SaveNewEntries takes a list of entries and adds any newer stats to the database
func (storeSaver *StoreSaver) SaveNewEntries(entries []models.CountryEntry) (newEntries []models.CountryEntry, err error) {
	newEntries, err = storeSaver.getNewRecords(entries)
	if err != nil {
		err = fmt.Errorf("failed to process Covid figures: %s", err.Error())
		return
	}

	if len(newEntries) > 0 {
		log.WithField("entries", len(newEntries)).Debug("adding new probe-19 data to the database")

		err = storeSaver.Store.Add(newEntries)
		if err != nil {
			err = fmt.Errorf("failed to add new entries in the database: %s", err.Error())
		}
	}
	return
}

func (storeSaver *StoreSaver) getNewRecords(entries []models.CountryEntry) (newEntries []models.CountryEntry, err error) {
	countries := getCountries(entries)

	var latest map[string]models.CountryEntry
	latest, err = storeSaver.Store.GetLatestForCountries(countries)

	for _, entry := range entries {
		latestEntry, found := latest[entry.Name]

		if !found || entry.Timestamp.After(latestEntry.Timestamp) {
			newEntries = append(newEntries, entry)
		}
	}

	return
}

func getCountries(entries []models.CountryEntry) (countries []string) {
	uniqueCountries := make(map[string]struct{})
	for _, entry := range entries {
		uniqueCountries[entry.Name] = struct{}{}
	}
	for name := range uniqueCountries {
		countries = append(countries, name)
	}
	sort.Strings(countries)
	return
}
