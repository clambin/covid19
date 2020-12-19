package probe

import (
	"covid19/internal/pushgateway"
	log "github.com/sirupsen/logrus"

	"covid19/internal/covid/apiclient"
	"covid19/internal/coviddb"
)

// Probe handle
type Probe struct {
	apiClient   apiclient.API
	db          coviddb.DB
	pushGateway *pushgateway.PushGateway
}

// NewProbe creates a new Probe handle
func NewProbe(apiClient apiclient.API, db coviddb.DB, pushGateway *pushgateway.PushGateway) *Probe {
	return &Probe{apiClient: apiClient, db: db, pushGateway: pushGateway}
}

// Run gets latest data, inserts any new entries in the DB and reports to Prometheus' pushGateway
func (probe *Probe) Run() error {
	countryStats, err := probe.apiClient.GetCountryStats()

	if err != nil {
		log.Warning(err)
	} else if len(countryStats) > 0 {
		log.Debugf("Got %d new entries", len(countryStats))

		newRecords, err := probe.findNewCountryStats(countryStats)

		if err == nil && len(newRecords) > 0 {
			log.Infof("Adding %d new entries", len(newRecords))

			err = probe.db.Add(newRecords)
		}

		if err == nil && probe.pushGateway != nil {
			countries := make([]string, 0, len(newRecords))
			for _, entry := range newRecords {
				countries = append(countries, entry.Name)
			}

			probe.pushGateway.Push(countries)
		}
	}

	return err
}

// findNewCountryStats returns any new stats (ie either more recent or the country has no entries yet)
func (probe *Probe) findNewCountryStats(newCountryStats map[string]apiclient.CountryStats) ([]coviddb.CountryEntry, error) {
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
