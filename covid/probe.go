package covid

import (
	"context"
	"fmt"
	"github.com/clambin/covid19/configuration"
	"github.com/clambin/covid19/covid/fetcher"
	"github.com/clambin/covid19/covid/saver"
	"github.com/clambin/covid19/covid/shoutrrr"
	"github.com/clambin/covid19/models"
	"github.com/clambin/go-common/set"
	"github.com/clambin/go-rapidapi"
	"golang.org/x/exp/slog"
	"time"
)

// Probe gets new COVID-19 stats for each country and, if they are new, adds them to the database
type Probe struct {
	fetcher.Fetcher
	saver.StoreSaver
	*Notifier
	invalidCountries set.Set[string]
}

const (
	rapidAPIHost = "covid-19-coronavirus-statistics.p.rapidapi.com"
)

// New creates a new Probe
func New(cfg *configuration.MonitorConfiguration, db saver.CovidAdderGetter) *Probe {
	var notifier *Notifier
	if cfg.Notifications.Enabled {
		router, err := shoutrrr.NewRouter(cfg.Notifications.URL)
		if err != nil {
			slog.Error("failed to create notification router", "err", err)
			panic(err)
		}
		notifier = &Notifier{
			Countries: set.Create(cfg.Notifications.Countries...),
			Sender:    router,
		}

	}
	return &Probe{
		Fetcher:          &fetcher.Client{API: rapidapi.New(rapidAPIHost, cfg.RapidAPIKey)},
		StoreSaver:       saver.StoreSaver{Store: db},
		Notifier:         notifier,
		invalidCountries: set.Create[string](),
	}
}

// Update gets new COVID-19 stats for each country and, if they are new, adds them to the database
func (p *Probe) Update(ctx context.Context) (int, error) {
	current, err := p.StoreSaver.Store.GetLatestForCountries(time.Time{})
	if err != nil {
		return 0, fmt.Errorf("get latest: %w", err)
	}

	countryStats, err := p.Fetcher.Fetch(ctx)
	if err == nil {
		countryStats, err = p.StoreSaver.SaveNewEntries(p.filterUnsupportedCountries(countryStats))
	}

	if err != nil {
		return 0, fmt.Errorf("update: %w", err)
	}

	if p.Notifier != nil {
		if err = p.Notify(current, countryStats); err != nil {
			slog.Error("failed to send notification", "err", err)
		}
	}
	return len(countryStats), nil
}

func (p *Probe) filterUnsupportedCountries(entries []models.CountryEntry) []models.CountryEntry {
	filteredEntries := make([]models.CountryEntry, 0, len(entries))
	for _, entry := range entries {
		code, found := CountryCodes[entry.Name]
		if !found {
			if !p.invalidCountries.Contains(entry.Name) {
				slog.Warn("unknown country name received from COVID-19 API", "name", entry.Name)
				p.invalidCountries.Add(entry.Name)
			}
			continue
		}
		entry.Code = code
		filteredEntries = append(filteredEntries, entry)
	}
	return filteredEntries
}
