package covidprobe

import (
	//"time"
	//"net/http"

	log "github.com/sirupsen/logrus"

	"covid19/internal/coviddb"
)

// CovidProbe handle
type CovidProbe struct {
	apiClient      *CovidAPIClient
	db              coviddb.CovidDB
	pushGateway     string
}

// NewCovidProbe creates a new CovidProbe handle
func NewCovidProbe(apiClient *CovidAPIClient, db coviddb.CovidDB, pushGateway string) (*CovidProbe) {
	return &CovidProbe{apiClient: apiClient, db: db, pushGateway: pushGateway}
}

// Run gets latest data, inserts any new entries in the DB and reports to Prometheus' pushGateway
func (probe *CovidProbe) Run() (error) {
	log.Debug("CovidProbe::Run{")
	// Call the API
	countryStats, err := probe.apiClient.GetCountryStats()

	if err == nil && len(countryStats) > 0 {
		log.Debugf("CovidProbe::Run : received %d entries", len(countryStats))

		dbRecords, err := probe.findNewCountryStats(countryStats)

		if err == nil {
			log.Debugf("CovidProbe::Run : will add  %d entries", len(dbRecords))

			// Add entries to DB
			err = probe.db.Add(dbRecords)
		}

		if err == nil && probe.pushGateway != "" {
			log.Debug("Call pushgateway here")
		}
	}

	if err != nil {
		log.Warning(err)
	}
	return err
}

// findNewCountryStats returns any new stats (ie either more recent or the country has no entries yet)
func (probe *CovidProbe) findNewCountryStats(newCountryStats map[string]Covid19CountryStats) ([]coviddb.CountryEntry, error) {
	result := make([]coviddb.CountryEntry, 0)

	lastDBEntries, err := probe.db.ListLatestByCountry()
	if err != nil { return result, err }

	for country, stats := range newCountryStats {
		// lastUpdate, ok := lastDBEntries[country]
		lastUpdate := lastDBEntries[country]
		ok := true
		if ok == false || stats.LastUpdate.After(lastUpdate) {
			code, ok := countryCodes[country]
			if ok == false {
				log.Warningf("unknown country '%s'. Skipping", country)
			} else {
				result = append(result, coviddb.CountryEntry{
						Timestamp: stats.LastUpdate,
						Code: code,
						Name: country,
						Confirmed: stats.Confirmed,
						Recovered: stats.Recovered,
						Deaths: stats.Deaths})
			}
		}
	}

	return result, nil
}
