package covid

import (
	"github.com/clambin/covid19/covid/fetcher"
	"github.com/clambin/covid19/models"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"time"
)

var (
	metricUpdates = prometheus.NewDesc(
		prometheus.BuildFQName("covid", "", "reported_count"),
		"New entries per country",
		[]string{"country"},
		nil,
	)
)

// Describe implements the prometheus collector Describe interface
func (probe *Probe) Describe(ch chan<- *prometheus.Desc) {
	ch <- metricUpdates
}

// Collect implements the prometheus collector Collect interface
func (probe *Probe) Collect(ch chan<- prometheus.Metric) {
	start := time.Now()
	probe.lock.RLock()
	defer probe.lock.RUnlock()

	for country := range fetcher.CountryCodes {
		count := probe.newUpdates[country]
		ch <- prometheus.MustNewConstMetric(metricUpdates, prometheus.CounterValue, float64(count), country)
	}
	log.WithField("duration", time.Since(start)).Debug("prometheus scrape done")
}

func (probe *Probe) setCountryUpdates(newEntries []models.CountryEntry) {
	probe.lock.Lock()
	defer probe.lock.Unlock()

	if probe.newUpdates == nil {
		probe.newUpdates = make(map[string]int)
	}

	for _, entry := range newEntries {
		count := probe.newUpdates[entry.Name]
		probe.newUpdates[entry.Name] = count + 1
	}
}
