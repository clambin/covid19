package fetcher

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	metricRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "covid_covid_api_requests_total",
		Help: "Number of COVID-19 API calls made",
	}, []string{"endpoint"})
	metricRequestErrorsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "covid_covid_api_request_errors_total",
		Help: "Number of failed COVID-19 API calls",
	}, []string{"endpoint"})
	metricRequestLatency = promauto.NewSummaryVec(prometheus.SummaryOpts{
		Name: "covid_covid_api_latency",
		Help: "Latency of COVID-19 API calls",
	}, []string{"endpoint"})
)
