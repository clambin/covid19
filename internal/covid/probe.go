package covid

import (
	"covid19/internal/coviddb"
	log "github.com/sirupsen/logrus"
)

// Probe handle
type Probe struct {
	apiClient   *APIClient
	db          coviddb.DB
	pushGateway *PushGateway
}

// NewProbe creates a new Probe handle
func NewProbe(apiClient *APIClient, db coviddb.DB, pushGateway *PushGateway) *Probe {
	return &Probe{apiClient: apiClient, db: db, pushGateway: pushGateway}
}

// Run gets latest data, inserts any new entries in the DB and reports to Prometheus' pushGateway
func (probe *Probe) Run() error {
	countryStats, err := probe.apiClient.GetCountryStats()

	if err == nil && len(countryStats) > 0 {
		log.Debugf("Got %d new entries", len(countryStats))

		dbRecords, err := probe.findNewCountryStats(countryStats)

		if err == nil && len(dbRecords) > 0 {
			log.Infof("Adding %d new entries", len(dbRecords))

			err = probe.db.Add(dbRecords)
		}

		if err == nil && probe.pushGateway != nil {
			countries := make([]string, 0, len(dbRecords))
			for _, entry := range dbRecords {
				countries = append(countries, entry.Name)
			}

			probe.pushGateway.Push(countries)
		}
	}

	if err != nil {
		log.Warning(err)
	}
	return err
}

// findNewCountryStats returns any new stats (ie either more recent or the country has no entries yet)
func (probe *Probe) findNewCountryStats(newCountryStats map[string]CountryStats) ([]coviddb.CountryEntry, error) {
	result := make([]coviddb.CountryEntry, 0)

	lastDBEntries, err := probe.db.ListLatestByCountry()
	if err != nil {
		return result, err
	}

	for country, stats := range newCountryStats {
		lastUpdate, ok := lastDBEntries[country]
		if ok == false || stats.LastUpdate.After(lastUpdate) {
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

	return result, nil
}
