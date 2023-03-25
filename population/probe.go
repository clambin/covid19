package population

import (
	"context"
	"github.com/clambin/covid19/covid"
	"golang.org/x/exp/slog"
	"golang.org/x/sync/semaphore"
)

// Probe downloads the latest population stats per country and stores them in the database
type Probe struct {
	APIClient
	store Adder
}

type Adder interface {
	Add(string, int64) error
}

// New creates a new Probe
func New(apiKey string, store Adder) *Probe {
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
			slog.Warn("unsupported country code received from population API. skipping", "code", code)
			continue
		}

		count++

		_ = maxJobs.Acquire(ctx, 1)
		go func(ctx context.Context, code, country string) {
			localError := probe.update(ctx, code, country)

			if localError != nil {
				slog.Error("failed to update population stats", "err", localError, "country", countryNames)
			}

			maxJobs.Release(1)
		}(ctx, code, country)
	}

	_ = maxJobs.Acquire(ctx, maxConcurrentJobs)

	return count, err
}

func countryCodes() (codes []string) {
	for _, code := range covid.CountryCodes {
		codes = append(codes, code)
	}
	return
}

func (probe *Probe) update(ctx context.Context, code, country string) (err error) {
	var population int64
	population, err = probe.APIClient.GetPopulation(ctx, country)

	if err == nil && population > 0 {
		slog.Debug("found population", "country", country, "population", population)
		err = probe.store.Add(code, population)
	}
	return
}
