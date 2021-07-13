package probe

import (
	"context"
	"github.com/clambin/covid19/coviddb"
	"github.com/clambin/covid19/population/db"
	"time"

	log "github.com/sirupsen/logrus"
)

// Probe handle
type Probe struct {
	APIClient
	popDB   db.DB
	covidDB coviddb.DB
}

// Create a new Probe handle
func Create(apiKey string, popDB db.DB, covidDB coviddb.DB) *Probe {
	return &Probe{
		APIClient: NewAPIClient(apiKey),
		popDB:     popDB,
		covidDB:   covidDB,
	}
}

// Run gets latest data and updates the database
func (probe *Probe) Run(ctx context.Context, interval time.Duration) (err error) {
	err = probe.runUpdate(ctx)
	ticker := time.NewTicker(interval)
loop:
	for err == nil {
		select {
		case <-ctx.Done():
			break loop
		case <-ticker.C:
			err = probe.runUpdate(ctx)
		}
	}
	ticker.Stop()
	return
}

type update struct {
	code    string
	country string
}

func (probe *Probe) runUpdate(ctx context.Context) (err error) {
	updater := NewUpdater(probe.update, 5)
	go updater.Run(ctx)

	start := time.Now()
	var codes []string
	codes, err = probe.covidDB.GetAllCountryCodes()
	if err == nil {
		for _, code := range codes {
			country, ok := countryNames[code]
			if !ok {
				log.WithField("code", code).Warning("unknown country code for population DB. skipping")
				continue
			}

			updater.Input <- &update{code: code, country: country}
		}

		updater.Stop <- struct{}{}
		<-updater.Done
		log.Infof("discovered %d country population figures in %v",
			len(codes), time.Now().Sub(start))
	}

	return err
}

func (probe *Probe) update(ctx context.Context, input interface{}) {
	newData := input.(*update)

	population, err := probe.APIClient.GetPopulation(ctx, newData.country)
	if err == nil {
		log.WithFields(log.Fields{
			"country":    newData.country,
			"population": population,
		}).Debug("found population")
		err = probe.popDB.Add(newData.code, population)
	}

	if err != nil {
		log.WithError(err).WithField("country", newData.country).Warning("could not get population stats")
	}
}
