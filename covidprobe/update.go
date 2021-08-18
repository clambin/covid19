package covidprobe

import (
	"context"
	"fmt"
	"github.com/clambin/covid19/coviddb"
	log "github.com/sirupsen/logrus"
	"sort"
	"time"
)

// Update retrieve the latest COVID-19 figures and inserts any new entries in the database
func (probe *Probe) Update(ctx context.Context) error {
	countryStats, err := probe.APIClient.GetCountryStats(ctx)
	if err != nil {
		return fmt.Errorf("failed to get Covid figures: " + err.Error())
	}

	log.WithField("entries", len(countryStats)).Debug("found covid-19 data")

	var newRecords []coviddb.CountryEntry
	newRecords, err = probe.getNewRecords(countryStats)
	if err != nil {
		return fmt.Errorf("failed to process Covid figures: %s", err.Error())
	}

	probe.recordUpdates(newRecords)

	if len(newRecords) > 0 {
		log.WithField("entries", len(newRecords)).Debug("adding new covid-19 data to the database")

		if probe.TestMode {
			// FIXME: mock only passes if the order of the elements is the same. Sort them here.
			sort.Slice(newRecords, func(i, j int) bool { return newRecords[i].Name < newRecords[j].Name })
		}

		err = probe.db.Add(newRecords)
		if err != nil {
			return fmt.Errorf("failed to add new entries in the database: %s", err.Error())
		}

		// TODO: shouldn't need to be triggered by probe. db to inform cache?
		if probe.cache != nil {
			probe.cache.Refresh()
		}

		if probe.Notifier != nil {
			_ = probe.Notifier.Notify(newRecords)
		}
	}

	return nil
}

// getNewRecords takes the newly collected country statistics and returns any new entries
func (probe *Probe) getNewRecords(newCountryStats map[string]CountryStats) (records []coviddb.CountryEntry, err error) {
	var latestUpdates map[string]time.Time
	latestUpdates, err = probe.db.ListLatestByCountry()

	for country, newStats := range newCountryStats {
		current, found := latestUpdates[country]

		if found == true && !newStats.LastUpdate.After(current) {
			// we already have a more recent entry in our db. skip this country
			continue
		}

		code, ok := CountryCodes[country]

		if ok == false {
			// country name is invalid.  log it, but only once
			if probe.knownInvalidCountries == nil {
				probe.knownInvalidCountries = make(map[string]struct{})
			}
			if _, reported := probe.knownInvalidCountries[country]; reported == false {
				log.WithField("country", country).Warning("skipping unknown country")
			}
			probe.knownInvalidCountries[country] = struct{}{}
			continue
		}

		records = append(records, coviddb.CountryEntry{
			Timestamp: newStats.LastUpdate,
			Code:      code,
			Name:      country,
			Confirmed: newStats.Confirmed,
			Recovered: newStats.Recovered,
			Deaths:    newStats.Deaths,
		})
	}
	return
}
