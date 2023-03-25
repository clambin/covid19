package saver

import (
	"fmt"
	"github.com/clambin/covid19/models"
	"golang.org/x/exp/slog"
	"time"
)

// StoreSaver saves new covid entries to the database
type StoreSaver struct {
	Store CovidAdderGetter
}

type CovidAdderGetter interface {
	Add([]models.CountryEntry) error
	GetLatestForCountries(time.Time) (map[string]models.CountryEntry, error)
}

// SaveNewEntries takes a list of entries and adds any newer stats to the database
func (s *StoreSaver) SaveNewEntries(entries []models.CountryEntry) ([]models.CountryEntry, error) {
	newEntries, err := s.getNewRecords(entries)
	if err != nil || len(newEntries) == 0 {
		return nil, err
	}
	slog.Debug("adding new probe-19 data to the database", "entries", len(newEntries))
	if err = s.Store.Add(newEntries); err != nil {
		err = fmt.Errorf("add: %w", err)
	}
	return newEntries, err
}

func (s *StoreSaver) getNewRecords(entries []models.CountryEntry) ([]models.CountryEntry, error) {
	latest, err := s.Store.GetLatestForCountries(time.Time{})
	if err != nil {
		return nil, err
	}

	var newEntries []models.CountryEntry
	for _, entry := range entries {
		latestEntry, found := latest[entry.Name]

		if !found || entry.Timestamp.After(latestEntry.Timestamp) {
			newEntries = append(newEntries, entry)
		}
	}
	return newEntries, err
}
