package covidprobe_test

import (
	"context"
	"github.com/clambin/covid19/configuration"
	"github.com/clambin/covid19/coviddb"
	dbMock "github.com/clambin/covid19/coviddb/mocks"
	"github.com/clambin/covid19/covidprobe"
	probeMock "github.com/clambin/covid19/covidprobe/mocks"
	"github.com/clambin/gotools/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestProbe_Describe(t *testing.T) {
	cfg := &configuration.MonitorConfiguration{}
	m := covidprobe.NewProbe(cfg, nil, nil)

	ch := make(chan *prometheus.Desc)
	go m.Describe(ch)

	for _, name := range []string{
		"covid_reported_count",
	} {
		metric := <-ch
		assert.Contains(t, metric.String(), "\""+name+"\"")
	}
}

func TestProbe_Collect(t *testing.T) {
	cfg := &configuration.MonitorConfiguration{
		Interval:      30 * time.Minute,
		RapidAPIKey:   configuration.ValueOrEnvVar{Value: "akey"},
		Notifications: configuration.NotificationConfiguration{},
	}
	db := &dbMock.DB{}
	apiClient := &probeMock.APIClient{}

	probe := covidprobe.NewProbe(cfg, db, nil)
	probe.APIClient = apiClient
	probe.TestMode = true

	timestamp := time.Now()

	// Initial update; nothing in database. Valid countries are inserted.
	apiClient.On("GetCountryStats", mock.Anything).Return(map[string]covidprobe.CountryStats{
		"Belgium":     {LastUpdate: timestamp, Confirmed: 40, Recovered: 10, Deaths: 1},
		"US":          {LastUpdate: timestamp, Confirmed: 20, Recovered: 15, Deaths: 5},
		"NotACountry": {LastUpdate: timestamp, Confirmed: 0, Recovered: 0, Deaths: 0},
	}, nil)
	db.On("ListLatestByCountry").Return(map[string]time.Time{}, nil).Once()
	db.On("Add", []coviddb.CountryEntry{
		{Code: "BE", Name: "Belgium", Timestamp: timestamp, Confirmed: 40, Recovered: 10, Deaths: 1},
		{Code: "US", Name: "US", Timestamp: timestamp, Confirmed: 20, Recovered: 15, Deaths: 5},
	}).Return(nil).Once()

	err := probe.Update(context.TODO())
	require.NoError(t, err)

	ch := make(chan prometheus.Metric)
	go probe.Collect(ch)

	todo := len(covidprobe.CountryCodes)
	for metric := range ch {
		todo--
		target := 0.0
		country := metrics.MetricLabel(metric, "country")
		if country == "Belgium" || country == "US" {
			target = 1.0
		}

		assert.Equal(t, target, metrics.MetricValue(metric).GetCounter().GetValue(), country)

		if todo == 0 {
			break
		}
	}

	db.On("ListLatestByCountry").Return(map[string]time.Time{
		"Belgium": timestamp,
		"US":      timestamp,
	}, nil).Once()
	err = probe.Update(context.TODO())
	require.NoError(t, err)

	ch = make(chan prometheus.Metric)
	go probe.Collect(ch)

	todo = len(covidprobe.CountryCodes)
	for metric := range ch {
		todo--
		target := 0.0
		country := metrics.MetricLabel(metric, "country")
		if country == "Belgium" || country == "US" {
			target = 1.0
		}

		assert.Equal(t, target, metrics.MetricValue(metric).GetCounter().GetValue(), country)

		if todo == 0 {
			break
		}
	}

	mock.AssertExpectationsForObjects(t, apiClient, db)
}
