package covidprobe

import (
	"context"
	"fmt"
	"github.com/clambin/covid19/configuration"
	"github.com/clambin/covid19/covidcache"
	"github.com/clambin/covid19/coviddb"
	"github.com/clambin/gotools/metrics"
	"github.com/containrrr/shoutrrr"
	"github.com/containrrr/shoutrrr/pkg/router"
	"github.com/containrrr/shoutrrr/pkg/types"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"time"
)

// Probe handle
type Probe struct {
	APIClient     APIClient
	db            coviddb.DB
	cache         *covidcache.Cache
	notifications *configuration.NotificationConfiguration
	notifier      *router.ServiceRouter

	NotifyCache           map[string]coviddb.CountryEntry
	KnownInvalidCountries map[string]bool
	PrometheusMetrics     map[string]int
}

// NewProbe creates a new Probe handle
func NewProbe(cfg *configuration.MonitorConfiguration, db coviddb.DB, cache *covidcache.Cache) (probe *Probe, err error) {
	probe = &Probe{
		APIClient:     NewAPIClient(cfg.RapidAPIKey.Value),
		db:            db,
		cache:         cache,
		notifications: &cfg.Notifications,
	}

	err = probe.initCache()

	if err == nil && probe.notifications.Enabled {
		var err2 error
		if probe.notifier, err2 = shoutrrr.CreateSender(cfg.Notifications.URL.Value); err2 != nil {
			log.WithField("err", err2).Error("failed to set up notifications")
			probe.notifier = nil
		}
	}

	return
}

// initCache initializes the NotifyCache structure
func (probe *Probe) initCache() error {
	var err error

	probe.NotifyCache = make(map[string]coviddb.CountryEntry)

	if probe.notifications.Enabled {
		for _, country := range probe.notifications.Countries {
			if code, ok := CountryCodes[country]; ok == true {
				var entry *coviddb.CountryEntry
				var found bool
				// Could add a db.GetLatest() though latest should always be less than 'now' on startup
				if entry, found, err = probe.db.GetLastBeforeDate(country, time.Now()); err == nil && found {
					probe.NotifyCache[country] = *entry
				} else {
					probe.NotifyCache[country] = coviddb.CountryEntry{
						Code: code,
						Name: country,
					}
				}
			} else {
				log.WithField("country", country).Warning("ignoring invalid country in notifications configuration")
			}
		}
	}

	return err
}

// Run gets latest data, inserts any new entries in the DB and reports to Prometheus' pushGateway
func (probe *Probe) Run(ctx context.Context, interval time.Duration) (err error) {
	err = probe.update(ctx)
	timer := time.NewTicker(interval)
loop:
	for err == nil {
		select {
		case <-ctx.Done():
			break loop
		case <-timer.C:
			err = probe.update(ctx)
		}
	}
	timer.Stop()
	return
}

func (probe *Probe) update(ctx context.Context) (err error) {
	var (
		countryStats map[string]CountryStats
		newRecords   = make([]coviddb.CountryEntry, 0)
	)

	countryStats, err = probe.APIClient.GetCountryStats(ctx)

	if err == nil {
		log.WithField("countryStats", len(countryStats)).Debug("covidProbe got new entries")

		newRecords, err = probe.getNewRecords(countryStats)
	}

	if err == nil {
		probe.metricsLatestUpdates(newRecords)

		if len(newRecords) > 0 {
			log.WithField("newRecords", len(newRecords)).Info("covidProbe inserting new entries")

			notifications := probe.getNotifications(newRecords)

			if err = probe.db.Add(newRecords); err != nil {
				log.WithField("err", err).Fatal("failed to add new entries in the DB")
			}

			if probe.cache != nil {
				probe.cache.Refresh()
			}

			if err = probe.sendNotifications(notifications); err != nil {
				log.WithField("key", err).Warn("failed to send notification")
			}
		}
	}
	return
}

// getNewRecords takes the newly collected country statistics and returns any new entries
func (probe *Probe) getNewRecords(newCountryStats map[string]CountryStats) ([]coviddb.CountryEntry, error) {
	var (
		err           error
		latestUpdates map[string]time.Time
	)
	records := make([]coviddb.CountryEntry, 0)

	latestUpdates, err = probe.db.ListLatestByCountry()

	for country, stats := range newCountryStats {
		current, ok := latestUpdates[country]

		// No entry for this country exists, or the new stats are more recent than what we have
		if ok == false || stats.LastUpdate.After(current) {
			var code string
			if code, ok = CountryCodes[country]; ok == false {
				if probe.KnownInvalidCountries == nil {
					probe.KnownInvalidCountries = make(map[string]bool)
				}
				if _, ok = probe.KnownInvalidCountries[country]; ok == false {
					log.WithField("country", country).Warning("skipping unknown country")
					probe.KnownInvalidCountries[country] = true
				}
			} else {
				records = append(records, coviddb.CountryEntry{
					Timestamp: stats.LastUpdate,
					Code:      code,
					Name:      country,
					Confirmed: stats.Confirmed,
					Recovered: stats.Recovered,
					Deaths:    stats.Deaths})
			}
		}
	}

	return records, err
}

type Notification struct {
	Title   string
	Message string
}

// getNotifications generates a notification for each new record for a country
func (probe *Probe) getNotifications(newEntries []coviddb.CountryEntry) []Notification {
	notifications := make([]Notification, 0)

	for _, newEntry := range newEntries {
		// NotifyCache only contains entries for countries we need to send notifications for
		if dbEntry, ok := probe.NotifyCache[newEntry.Name]; ok {
			notifications = append(notifications, Notification{
				Title: "New covid data for " + newEntry.Name,
				Message: fmt.Sprintf("Confirmed: %d, deaths: %d, recovered: %d",
					newEntry.Confirmed-dbEntry.Confirmed,
					newEntry.Deaths-dbEntry.Deaths,
					newEntry.Recovered-dbEntry.Recovered,
				),
			})

			probe.NotifyCache[newEntry.Name] = newEntry
		}
	}

	return notifications
}

// sendNotifications sends a notification for each country that has a new update
func (probe *Probe) sendNotifications(notifications []Notification) error {
	var errs []error

	for _, notification := range notifications {
		params := types.Params{}
		params.SetTitle(notification.Title)
		errs = probe.notifier.Send(notification.Message, &params)
		for _, e := range errs {
			if e != nil {
				return e
			}
		}
	}
	return nil
}

// Metrics to be reported
var (
	reportedCount = metrics.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "covid_reported_count",
			Help: "New entries per country",
		},
		[]string{
			"country",
		},
	)
)

func (probe *Probe) metricsLatestUpdates(newEntries []coviddb.CountryEntry) {
	if probe.PrometheusMetrics == nil {
		probe.PrometheusMetrics = make(map[string]int)

		for name := range CountryCodes {
			probe.PrometheusMetrics[name] = 0
		}
	}

	for country := range probe.PrometheusMetrics {
		probe.PrometheusMetrics[country] = 0
	}

	for _, newEntry := range newEntries {
		count, _ := probe.PrometheusMetrics[newEntry.Name]
		probe.PrometheusMetrics[newEntry.Name] = count + 1
	}
	for country, count := range probe.PrometheusMetrics {
		reportedCount.WithLabelValues(country).Set(float64(count))
	}
}
