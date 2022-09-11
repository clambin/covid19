package covid

import (
	"context"
	"fmt"
	"github.com/clambin/covid19/configuration"
	"github.com/clambin/covid19/covid/fetcher"
	"github.com/clambin/covid19/covid/notifier"
	"github.com/clambin/covid19/covid/saver"
	"github.com/clambin/covid19/db"
	"github.com/clambin/go-rapidapi"
	log "github.com/sirupsen/logrus"
	"sync"
)

// Probe gets new COVID-19 stats for each country and, if they are new, adds them to the database
type Probe struct {
	fetcher.Fetcher
	saver.Saver
	notifier.Notifier
	newUpdates map[string]int
	lock       sync.RWMutex
}

const (
	rapidAPIHost = "covid-19-coronavirus-statistics.p.rapidapi.com"
)

// New creates a new Probe
func New(cfg *configuration.MonitorConfiguration, db db.CovidStore) *Probe {
	var n notifier.Notifier
	if cfg.Notifications.Enabled {
		n = notifier.NewNotifier(
			notifier.NewNotificationSender(cfg.Notifications.URL),
			cfg.Notifications.Countries,
			db)
	}
	return &Probe{
		Fetcher: &fetcher.Client{
			API: rapidapi.New(rapidAPIHost, cfg.RapidAPIKey),
		},
		Saver:    &saver.StoreSaver{Store: db},
		Notifier: n,
	}
}

// Update gets new COVID-19 stats for each country and, if they are new, adds them to the database
func (probe *Probe) Update(ctx context.Context) (int, error) {
	countryStats, err := probe.Fetcher.GetCountryStats(ctx)
	if err == nil {
		countryStats, err = probe.Saver.SaveNewEntries(countryStats)
	}

	if err != nil {
		return 0, fmt.Errorf("failed to get Covid figures: " + err.Error())
	}

	if probe.Notifier != nil {
		if err = probe.Notifier.Notify(countryStats); err != nil {
			log.WithError(err).Error("failed to send notification")
		}
	}

	probe.setCountryUpdates(countryStats)

	return len(countryStats), nil
}
