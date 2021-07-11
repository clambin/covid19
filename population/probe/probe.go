package probe

import (
	"github.com/clambin/covid19/coviddb"
	"github.com/clambin/covid19/population/db"

	log "github.com/sirupsen/logrus"
)

// Probe handle
type Probe struct {
	APIClient
	popDB   db.DB
	covidDB coviddb.DB
}

// Create a new Probe handle
func Create(apiKey string, popdb db.DB, coviddb coviddb.DB) *Probe {
	return &Probe{
		APIClient: NewAPIClient(apiKey),
		popDB:     popdb,
		covidDB:   coviddb,
	}
}

// Run gets latest data and updates the database
func (probe *Probe) Run() (err error) {
	var codes []string
	codes, err = probe.covidDB.GetAllCountryCodes()

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
		log.WithFields(log.Fields{"country": country, "code": code}).Debug("looking up population")

		population, err = probe.APIClient.GetPopulation(country)
		if err == nil {
			log.WithField("population", population).Debug("found population")
			err = probe.popDB.Add(code, population)
		} else {
			log.WithError(err).WithField("country", country).Warning("could not get population stats")
			err = nil
		}
	}

	return err
}
