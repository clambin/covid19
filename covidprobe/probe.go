package covidprobe

import (
	"github.com/clambin/covid19/configuration"
	"github.com/clambin/covid19/covidcache"
	"github.com/clambin/covid19/coviddb"
	"github.com/prometheus/client_golang/prometheus"
	"sync"
)

// Probe handle
type Probe struct {
	APIClient
	// If TestMode is set, Probe will sort entries before adding them to the database. Used for mocking
	TestMode bool
	Notifier *Notifier
	db       coviddb.DB
	cache    *covidcache.Cache

	knownInvalidCountries map[string]struct{}

	lock          sync.RWMutex
	newUpdates    map[string]int
	metricUpdates *prometheus.Desc
}

// NewProbe creates a new Probe
func NewProbe(cfg *configuration.MonitorConfiguration, db coviddb.DB, cache *covidcache.Cache) (probe *Probe) {
	var notifier *Notifier
	if cfg.Notifications.Enabled {
		notifier = NewNotifier(
			newNotificationSender(cfg.Notifications.URL.Get()),
			cfg.Notifications.Countries,
			db,
		)
	}

	probe = &Probe{
		APIClient:  NewAPIClient(cfg.RapidAPIKey.Value),
		db:         db,
		cache:      cache,
		Notifier:   notifier,
		newUpdates: make(map[string]int),
		metricUpdates: prometheus.NewDesc(
			prometheus.BuildFQName("covid", "", "reported_count"),
			"New entries per country",
			[]string{"country"},
			nil,
		),
	}

	return
}
