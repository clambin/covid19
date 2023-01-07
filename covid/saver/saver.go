package saver

import (
	"fmt"
	"github.com/clambin/covid19/db"
	"github.com/clambin/covid19/models"
	"golang.org/x/exp/slog"
)

// Saver stores new entries in the database
//
//go:generate mockery --name Saver
type Saver interface {
	SaveNewEntries(entries []models.CountryEntry) (newEntries []models.CountryEntry, err error)
}

// StoreSaver implements the Saver interface for CovidStore
type StoreSaver struct {
	Store db.CovidStore
}

var _ Saver = &StoreSaver{}

// SaveNewEntries takes a list of entries and adds any newer stats to the database
func (storeSaver *StoreSaver) SaveNewEntries(entries []models.CountryEntry) ([]models.CountryEntry, error) {
	newEntries, err := storeSaver.getNewRecords(entries)
	if err == nil && len(newEntries) > 0 {
		slog.Debug("adding new probe-19 data to the database", "entries", len(newEntries))
		err = storeSaver.Store.Add(newEntries)
		if err != nil {
			err = fmt.Errorf("add: %w", err)
		}
	}
	return newEntries, err
}

func (storeSaver *StoreSaver) getNewRecords(entries []models.CountryEntry) (newEntries []models.CountryEntry, err error) {
	var latest map[string]models.CountryEntry
	latest, err = storeSaver.Store.GetLatestForCountries()

	for _, entry := range entries {
		latestEntry, found := latest[entry.Name]

		if !found || entry.Timestamp.After(latestEntry.Timestamp) {
			newEntries = append(newEntries, entry)
		}
	}

	return
}
