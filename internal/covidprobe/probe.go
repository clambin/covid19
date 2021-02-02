package covidprobe

import (
	"covid19/internal/configuration"
	"covid19/internal/coviddb"
	"fmt"
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
	notifications *configuration.NotificationsConfiguration
	lastUpdate    map[string]time.Time
	notifier      *router.ServiceRouter
}

// NewProbe creates a new Probe handle
func NewProbe(cfg *configuration.MonitorConfiguration, db coviddb.DB) *Probe {
	probe := Probe{
		APIClient:     NewAPIClient(cfg.RapidAPIKey),
		db:            db,
		notifications: &cfg.Notifications,
	}

	if cfg.Notifications.Enabled {
		if len(cfg.Notifications.Countries) == 0 {
			log.Warning("notifications enabled but no countries specified. Ignoring")
		} else {
			var err error
			if probe.notifier, err = shoutrrr.CreateSender(cfg.Notifications.URL); err == nil {
				err = probe.cacheLatestUpdates()
			}

			if err != nil {
				log.WithField("err", err).Warning("failed to set up notifications")
				probe.notifier = nil
			}
		}
	}

	return &probe
}

// Run gets latest data, inserts any new entries in the DB and reports to Prometheus' pushGateway
func (probe *Probe) Run() error {
	var (
		err          error
		countryStats map[string]CountryStats
		newRecords   = make([]coviddb.CountryEntry, 0)
	)

	countryStats, err = probe.APIClient.GetCountryStats()

	if err == nil && len(countryStats) > 0 {
		log.WithField("countryStats", len(countryStats)).Debug("covidProbe got new entries")

		newRecords, err = probe.findNewCountryStats(countryStats)
	}

	if err == nil && len(newRecords) > 0 {
		log.WithField("newRecords", len(newRecords)).Info("covidProbe inserting new entries")

		/*
			if err = probe.cacheLatestUpdates(); err != nil {
				log.WithField("err", err).Warning("failed to get latest entries in DB")
			}
		*/

		probe.metricsLatestUpdates(newRecords)

		if probe.notifier != nil {
			err = probe.notifyLatestUpdates(newRecords)
			log.WithFields(log.Fields{
				"err": err,
				"url": probe.notifications.URL,
			}).Debug("notification failed")
		}

		if err = probe.db.Add(newRecords); err != nil {
			log.WithField("err", err).Fatal("failed to add new entries in the DB")
		}
	}

	return err
}

// findNewCountryStats returns any new stats (ie either more recent or the country has no entries yet)
func (probe *Probe) findNewCountryStats(newCountryStats map[string]CountryStats) ([]coviddb.CountryEntry, error) {
	result := make([]coviddb.CountryEntry, 0)

	latestDBEntries, err := probe.db.ListLatestByCountry()

	if err == nil {
		for country, stats := range newCountryStats {
			latestUpdate, ok := latestDBEntries[country]
			if ok == false || stats.LastUpdate.After(latestUpdate) {
				code, ok := countryCodes[country]
				if ok == false {
					log.WithField("country", country).Warning("skipping unknown country")
				} else {
					result = append(result, coviddb.CountryEntry{
						Timestamp: stats.LastUpdate,
						Code:      code,
						Name:      country,
						Confirmed: stats.Confirmed,
						Recovered: stats.Recovered,
						Deaths:    stats.Deaths})
				}
			}
		}
	}

	return result, nil
}

// cacheLatestUpdates gets the last time for all countries we need to report on
func (probe *Probe) cacheLatestUpdates() error {
	var (
		err           error
		latestUpdates map[string]time.Time
	)

	probe.lastUpdate = make(map[string]time.Time)

	if latestUpdates, err = probe.db.ListLatestByCountry(); err == nil {
		for _, country := range probe.notifications.Countries {
			if lastUpdate, ok := latestUpdates[country]; ok {
				probe.lastUpdate[country] = lastUpdate
			}
		}
	}

	return err
}

// shouldNotify helper function to check if we should send a notification when we receive new data for a country
func (probe *Probe) shouldNotify(entryCountry string) bool {
	for _, country := range probe.notifications.Countries {
		if country == entryCountry {
			return true
		}
	}
	return false
}

// notifyLatestUpdates sends a notification for each country that has a new update
func (probe *Probe) notifyLatestUpdates(newEntries []coviddb.CountryEntry) error {
	var (
		err     error
		dbEntry *coviddb.CountryEntry
	)

	for _, newEntry := range newEntries {
		// Do we need to send a notification?
		if probe.shouldNotify(newEntry.Name) {
			oldTime, ok := probe.lastUpdate[newEntry.Name]

			if ok == false || newEntry.Timestamp.After(oldTime) {
				// get previous values
				if dbEntry, err = probe.db.GetLastBeforeDate(newEntry.Name, newEntry.Timestamp); err == nil && dbEntry != nil {
					// send notification
					// FIXME: how to use shoutrrr during unit testing?
					params := types.Params{}
					params.SetTitle("New covid data for " + newEntry.Name)
					err2 := probe.notifier.Send(
						fmt.Sprintf("New confirmed: %d\nNew deaths: %d\nNew recovered: %d",
							newEntry.Confirmed-dbEntry.Confirmed,
							newEntry.Deaths-dbEntry.Deaths,
							newEntry.Recovered-dbEntry.Recovered,
						),
						&params,
					)
					log.WithFields(log.Fields{
						"err":     err2,
						"country": newEntry.Name,
					}).Debug("notification sent")
				}
				probe.lastUpdate[newEntry.Name] = newEntry.Timestamp
			}
		}
	}

	return err
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
	summary := make(map[string]int)
	for _, newEntry := range newEntries {
		if count, ok := summary[newEntry.Name]; ok == false {
			summary[newEntry.Name] = 1
		} else {
			summary[newEntry.Name] = count + 1
		}
	}
	for country, count := range summary {
		reportedCount.WithLabelValues(country).Set(float64(count))
	}
}
