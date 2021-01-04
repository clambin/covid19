package reporters

import (
	"github.com/clambin/gotools/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
	log "github.com/sirupsen/logrus"

	"covid19/internal/coviddb"
)

// UpdatesReporter reports each country getting new data to PushGateway
type CountriesReporter struct {
	pusher *push.Pusher
}

// NewUpdatesReporter creates a new UpdatesReporter with the specified PushGateway URL
func NewCountriesReporter(url string) *CountriesReporter {
	registry := prometheus.NewRegistry()
	registry.MustRegister(reportedCount)
	pusher := push.New(url, "covid19mon").Gatherer(registry)

	return &CountriesReporter{pusher: pusher}
}

// Report acts on new covidprobe entries
func (reporter *CountriesReporter) Report(entries []coviddb.CountryEntry) {
	countries := make(map[string]int, len(entries))
	for _, entry := range entries {
		if _, ok := countries[entry.Name]; ok == false {
			countries[entry.Name] = 0
		}
		countries[entry.Name] += 1
	}
	log.Debugf("Sending metrics for %d countries", len(countries))
	for country, value := range countries {
		reportedCount.WithLabelValues(country).Set(float64(value))
	}
	err := reporter.pusher.Push()
	if err != nil {
		log.Warningf("Could not push metrics: %v", err)
	}
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
