package probe

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Prometheus metrics
var (
	metricRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "covid_population_api_requests_total",
		Help: "Number of population API calls made",
	}, []string{"endpoint"})
	metricRequestErrorsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "covid_population_api_request_errors_total",
		Help: "Number of failed population API calls",
	}, []string{"endpoint"})
	metricRequestLatency = promauto.NewSummaryVec(prometheus.SummaryOpts{
		Name: "covid_population_api_latency",
		Help: "Latency of population API calls",
	}, []string{"endpoint"})
)
