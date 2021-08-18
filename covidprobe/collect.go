package covidprobe

import (
	"github.com/clambin/covid19/coviddb"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"time"
)

// Describe implements the prometheus collector Describe interface
func (probe *Probe) Describe(ch chan<- *prometheus.Desc) {
	ch <- probe.metricUpdates
}

// Collect implements the prometheus collector Collect interface
func (probe *Probe) Collect(ch chan<- prometheus.Metric) {
	start := time.Now()
	probe.lock.RLock()
	defer probe.lock.RUnlock()

	for country := range CountryCodes {
		count, _ := probe.newUpdates[country]
		ch <- prometheus.MustNewConstMetric(probe.metricUpdates, prometheus.CounterValue, float64(count), country)
	}
	log.WithField("duration", time.Now().Sub(start)).Debug("prometheus scrape done")
}

func (probe *Probe) recordUpdates(newEntries []coviddb.CountryEntry) {
	probe.lock.Lock()
	defer probe.lock.Unlock()

	for _, entry := range newEntries {
		count, _ := probe.newUpdates[entry.Name]
		probe.newUpdates[entry.Name] = count + 1
	}
}
