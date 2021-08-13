package probe

import (
	"context"
	"github.com/clambin/covid19/coviddb"
	"github.com/clambin/covid19/population/db"
	"golang.org/x/sync/semaphore"
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
	probe.runUpdate(ctx)
	ticker := time.NewTicker(interval)

	for running := true; running; {
		select {
		case <-ctx.Done():
			running = false
		case <-ticker.C:
			probe.runUpdate(ctx)
		}
	}

	ticker.Stop()
	return
}

func (probe *Probe) runUpdate(ctx context.Context) {
	const maxConcurrentJobs = 5
	start := time.Now()
	var codes []string
	var err error
	codes, err = probe.covidDB.GetAllCountryCodes()
	if err != nil {
		return
	}

	maxJobs := semaphore.NewWeighted(maxConcurrentJobs)
	for _, code := range codes {
		country, ok := countryNames[code]

		if ok == false {
			log.WithField("code", code).Warning("unknown country code found in covid DB. skipping")
			continue
		}

		_ = maxJobs.Acquire(ctx, 1)
		go func(ctx context.Context, code, country string) {
			localError := probe.update(ctx, code, country)

			if localError != nil {
				log.WithError(localError).Errorf("failed to update population stats for %s", country)
			}

			maxJobs.Release(1)
		}(ctx, code, country)
	}

	_ = maxJobs.Acquire(ctx, maxConcurrentJobs)

	log.Infof("discovered %d country population figures in %v", len(codes), time.Now().Sub(start))
}

func (probe *Probe) update(ctx context.Context, code, country string) (err error) {
	var population int64
	population, err = probe.APIClient.GetPopulation(ctx, country)

	if err == nil {
		log.WithFields(log.Fields{"country": country, "population": population}).Debug("found population")
		err = probe.popDB.Add(code, population)
	}
	return
}
