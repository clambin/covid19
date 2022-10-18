package population

import (
	"context"
	"github.com/clambin/covid19/covid/fetcher"
	"github.com/clambin/covid19/db"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/semaphore"
)

// Probe downloads the latest population stats per country and stores them in the database
type Probe struct {
	APIClient
	store db.PopulationStore
}

// New creates a new Probe
func New(apiKey string, store db.PopulationStore) *Probe {
	return &Probe{
		APIClient: NewAPIClient(apiKey),
		store:     store,
	}
}

const maxConcurrentJobs = 5

// Update gets the current population for each supported country and stores it in the database
func (probe *Probe) Update(ctx context.Context) (count int, err error) {
	maxJobs := semaphore.NewWeighted(maxConcurrentJobs)
	for _, code := range countryCodes() {
		country, found := countryNames[code]

		if !found {
			log.WithField("code", code).Warning("unsupported country code received from population API. skipping")
			continue
		}

		count++

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

	return count, err
}

func countryCodes() (codes []string) {
	for _, code := range fetcher.CountryCodes {
		codes = append(codes, code)
	}
	return
}

func (probe *Probe) update(ctx context.Context, code, country string) (err error) {
	var population int64
	population, err = probe.APIClient.GetPopulation(ctx, country)

	if err == nil && population > 0 {
		log.WithFields(log.Fields{"country": country, "population": population}).Debug("found population")
		err = probe.store.Add(code, population)
	}
	return
}