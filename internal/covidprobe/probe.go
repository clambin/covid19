package covidprobe

import (
	log "github.com/sirupsen/logrus"

	"covid19/internal/coviddb"
	"covid19/internal/reporters"
)

// Probe handle
type Probe struct {
	APIClient APIClient
	db        coviddb.DB
	reporters *reporters.Reporters
}

// NewProbe creates a new Probe handle
func NewProbe(apiKey string, db coviddb.DB, reporters *reporters.Reporters) *Probe {
	return &Probe{
		APIClient: NewAPIClient(apiKey),
		db:        db,
		reporters: reporters,
	}
}

// Run gets latest data, inserts any new entries in the DB and reports to Prometheus' pushGateway
func (probe *Probe) Run() error {
	countryStats, err := probe.APIClient.GetCountryStats()

	if err != nil {
		log.Warning(err)
	} else if len(countryStats) > 0 {
		log.Debugf("Got %d new entries", len(countryStats))

		newRecords, err := probe.findNewCountryStats(countryStats)

		if err == nil && len(newRecords) > 0 {
			if probe.reporters != nil {
				probe.reporters.Report(newRecords)
			}

			log.Infof("Adding %d new entries", len(newRecords))
			err = probe.db.Add(newRecords)
		}
	}

	return err
}

// findNewCountryStats returns any new stats (ie either more recent or the country has no entries yet)
func (probe *Probe) findNewCountryStats(newCountryStats map[string]CountryStats) ([]coviddb.CountryEntry, error) {
	result := make([]coviddb.CountryEntry, 0)

	latestDBEntries, err := probe.db.ListLatestByCountry()

	if err == nil {
		for country, stats := range newCountryStats {
			latestUpdate, ok := latestDBEntries[country]
			if ok == false || stats.LastUpdate.After(latestUpdate) {
				code, ok := countryCodes[country]
				if ok == false {
					log.Warningf("unknown country '%s'. Skipping", country)
				} else {
					result = append(result, coviddb.CountryEntry{
						Timestamp: stats.LastUpdate,
						Code:      code,
						Name:      country,
						Confirmed: stats.Confirmed,
						Recovered: stats.Recovered,
						Deaths:    stats.Deaths})
				}
			}
		}
	}

	return result, nil
}
