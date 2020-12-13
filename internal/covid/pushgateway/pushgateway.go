package pushgateway

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"

	log "github.com/sirupsen/logrus"
)

// PushGateway handle
type PushGateway struct {
	pusher *push.Pusher
}

// NewPushGateway creates a new PushGateway handle
func NewPushGateway(url string) *PushGateway {
	if url == "" {
		return nil
	}

	registry := prometheus.NewRegistry()
	registry.MustRegister(reportedCount)
	pusher := push.New(url, "covid19mon").Gatherer(registry)

	return &PushGateway{pusher: pusher}
}

// Push reports the number of new records
func (pushGateway *PushGateway) Push(countries []string) {
	log.Debugf("Sending metrics for %d countries", len(countries))
	for _, country := range countries {
		reportedCount.WithLabelValues(country).Set(float64(1))
	}
	err := pushGateway.pusher.Push()
	if err != nil {
		log.Warningf("Could not push metrics: %v", err)
	}
}

// Metrics to be reported
var (
	reportedCount = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "covid_reported_count",
			Help: "New entries per country",
		},
		[]string{
			"country",
		},
	)
)
