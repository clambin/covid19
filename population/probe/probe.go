package probe

import (
	coviddb2 "github.com/clambin/covid19/coviddb"
	db2 "github.com/clambin/covid19/population/db"

	log "github.com/sirupsen/logrus"
)

// Probe handle
type Probe struct {
	APIClient APIClient
	popdb     db2.DB
	coviddb   coviddb2.DB
}

// Create a new Probe handle
func Create(apiKey string, popdb db2.DB, coviddb coviddb2.DB) *Probe {
	return &Probe{
		APIClient: NewAPIClient(apiKey),
		popdb:     popdb,
		coviddb:   coviddb,
	}
}

// Run gets latest data and updates the database
func (probe *Probe) Run() (err error) {
	var codes []string
	codes, err = probe.coviddb.GetAllCountryCodes()

	if err != nil {
		return
	}

	var population int64
	for _, code := range codes {
		country, ok := countryNames[code]
		if !ok {
			log.WithField("code", code).Warning("unknown country code for population DB. skipping")
			continue
		}

		population, err = probe.APIClient.GetPopulation(country)
		if err == nil {
			err = probe.popdb.Add(code, population)
		} else {
			log.WithError(err).WithField("country", country).Warning("could not get population stats")
			err = nil
		}
	}

	return err
}
