package probe

import (
	"context"
	"fmt"
	"github.com/clambin/covid19/configuration"
	"github.com/clambin/covid19/covid/probe/fetcher"
	"github.com/clambin/covid19/covid/probe/notifier"
	"github.com/clambin/covid19/covid/probe/saver"
	"github.com/clambin/covid19/covid/store"
	"github.com/clambin/covid19/models"
	"github.com/clambin/gotools/rapidapi"
	log "github.com/sirupsen/logrus"
	"net/http"
	"sync"
	"time"
)

// Covid19Probe gets new COVID-19 stats for each country and, if they are new, adds them to the database
type Covid19Probe struct {
	fetcher.Fetcher
	saver.Saver
	notifier.Notifier
	newUpdates map[string]int
	lock       sync.RWMutex
}

const (
	rapidAPIHost = "covid-19-coronavirus-statistics.p.rapidapi.com"
)

// New creates a new Covid19Probe
func New(cfg *configuration.MonitorConfiguration, db store.CovidStore) *Covid19Probe {
	var n notifier.Notifier
	if cfg.Notifications.Enabled {
		n = notifier.NewNotifier(
			notifier.NewNotificationSender(cfg.Notifications.URL.Get()),
			cfg.Notifications.Countries,
			db)
	}
	return &Covid19Probe{
		Fetcher: &fetcher.Client{
			API: &rapidapi.Client{
				HTTPClient: &http.Client{},
				APIKey:     cfg.RapidAPIKey.Get(),
				Hostname:   rapidAPIHost,
			},
		},
		Saver:    &saver.StoreSaver{Store: db},
		Notifier: n,
	}
}

// Update gets new COVID-19 stats for each country and, if they are new, adds them to the database
func (probe *Covid19Probe) Update(ctx context.Context) (err error) {
	start := time.Now()

	var countryStats []models.CountryEntry
	countryStats, err = probe.Fetcher.GetCountryStats(ctx)
	if err == nil {
		log.WithField("entries", len(countryStats)).Info("found covid-19 data")
		countryStats, err = probe.Saver.SaveNewEntries(countryStats)
		log.WithField("entries", len(countryStats)).Info("saved covid-19 data")
	}

	if err != nil {
		return fmt.Errorf("failed to get Covid figures: " + err.Error())
	}

	if probe.Notifier != nil {
		err2 := probe.Notifier.Notify(countryStats)
		if err2 != nil {
			log.WithError(err2).Error("failed to send notification")
		}
	}

	probe.setCountryUpdates(countryStats)

	log.Infof("discovered %d country figures in %v", len(countryStats), time.Since(start))

	return
}
