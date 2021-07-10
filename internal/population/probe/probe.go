package probe

import (
	"github.com/clambin/covid19/internal/coviddb"
	"github.com/clambin/covid19/internal/population/db"

	log "github.com/sirupsen/logrus"
)

// Probe handle
type Probe struct {
	APIClient APIClient
	popdb     db.DB
	coviddb   coviddb.DB
}

// Create a new Probe handle
func Create(apiKey string, popdb db.DB, coviddb coviddb.DB) *Probe {
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
