package probe_test

import (
	"context"
	"fmt"
	"github.com/clambin/covid19/configuration"
	"github.com/clambin/covid19/covid/probe"
	"github.com/clambin/covid19/covid/probe/fetcher"
	mockFetcher "github.com/clambin/covid19/covid/probe/fetcher/mocks"
	mockNotifier "github.com/clambin/covid19/covid/probe/notifier/mocks"
	mockSaver "github.com/clambin/covid19/covid/probe/saver/mocks"
	mockCovidStore "github.com/clambin/covid19/covid/store/mocks"
	"github.com/clambin/covid19/models"
	"github.com/clambin/go-metrics/tools"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestCovid19Probe_Update(t *testing.T) {
	cfg := &configuration.MonitorConfiguration{
		RapidAPIKey: configuration.ValueOrEnvVar{
			Value: "1234",
		},
		Notifications: configuration.NotificationConfiguration{
			Enabled:   true,
			Countries: []string{"Belgium", "US"},
		},
	}
	db := &mockCovidStore.CovidStore{}
	timeStamp := time.Now()
	db.
		On("GetLatestForCountries", []string{"Belgium", "US"}).
		Return(
			map[string]models.CountryEntry{
				"Belgium": {Timestamp: timeStamp, Name: "Belgium", Code: "BE", Confirmed: 10, Deaths: 2, Recovered: 1},
				"US":      {Timestamp: timeStamp, Name: "US", Code: "US", Confirmed: 100, Deaths: 20, Recovered: 10},
			},
			nil,
		)
	p := probe.New(cfg, db)

	f := &mockFetcher.Fetcher{}
	n := &mockNotifier.Notifier{}
	s := &mockSaver.Saver{}

	p.Fetcher = f
	p.Saver = s
	p.Notifier = n

	countryStats := []models.CountryEntry{
		{Timestamp: timeStamp.Add(-24 * time.Hour), Name: "Belgium", Code: "BE", Confirmed: 8, Deaths: 1, Recovered: 1},
		{Timestamp: timeStamp.Add(24 * time.Hour), Name: "US", Code: "US", Confirmed: 120, Deaths: 25, Recovered: 15},
	}
	newCountryStats := []models.CountryEntry{{Timestamp: timeStamp.Add(24 * time.Hour), Name: "US", Code: "US", Confirmed: 120, Deaths: 25, Recovered: 15}}

	f.
		On("GetCountryStats", mock.AnythingOfType("*context.emptyCtx")).
		Return(countryStats, nil).
		Once()
	s.
		On("SaveNewEntries", countryStats).
		Return(newCountryStats, nil).
		Once()
	n.
		On("Notify", newCountryStats).
		Return(nil).
		Once()

	err := p.Update(context.Background())
	require.NoError(t, err)

	ch := make(chan prometheus.Metric)
	go p.Collect(ch)

	for i := len(fetcher.CountryCodes); i > 0; i-- {
		metric := <-ch
		country := tools.MetricLabel(metric, "country")
		target := 0.0
		if country == "US" {
			target = 1.0
		}

		assert.Equal(t, target, tools.MetricValue(metric).GetCounter().GetValue(), country)
	}

	mock.AssertExpectationsForObjects(t, db, s, f, n)
}

func TestCovid19Probe_Update_Errors(t *testing.T) {
	f := &mockFetcher.Fetcher{}
	s := &mockSaver.Saver{}
	n := &mockNotifier.Notifier{}

	p := probe.Covid19Probe{
		Fetcher:  f,
		Saver:    s,
		Notifier: n,
	}

	f.
		On("GetCountryStats", mock.AnythingOfType("*context.emptyCtx")).
		Return(nil, fmt.Errorf("unable to get new country stats")).
		Once()

	err := p.Update(context.Background())
	require.Error(t, err)

	timeStamp := time.Now()
	countryStats := []models.CountryEntry{
		{Timestamp: timeStamp.Add(-24 * time.Hour), Name: "Belgium", Code: "BE", Confirmed: 8, Deaths: 1, Recovered: 1},
		{Timestamp: timeStamp.Add(24 * time.Hour), Name: "US", Code: "US", Confirmed: 120, Deaths: 25, Recovered: 15},
	}
	newCountryStats := []models.CountryEntry{{Timestamp: timeStamp.Add(24 * time.Hour), Name: "US", Code: "US", Confirmed: 120, Deaths: 25, Recovered: 15}}

	f.
		On("GetCountryStats", mock.AnythingOfType("*context.emptyCtx")).
		Return(countryStats, nil).
		Once()
	s.
		On("SaveNewEntries", countryStats).
		Return(nil, fmt.Errorf("unable to store new entries")).
		Once()

	err = p.Update(context.Background())
	require.Error(t, err)

	f.
		On("GetCountryStats", mock.AnythingOfType("*context.emptyCtx")).
		Return(countryStats, nil).
		Once()
	s.
		On("SaveNewEntries", countryStats).
		Return(newCountryStats, nil).
		Once()
	n.
		On("Notify", newCountryStats).
		Return(fmt.Errorf("unable to send notifications")).
		Once()

	err = p.Update(context.Background())
	require.NoError(t, err)

	mock.AssertExpectationsForObjects(t, f, s, n)
}

func TestCovid19Probe_Describe(t *testing.T) {
	p := probe.Covid19Probe{}
	ch := make(chan *prometheus.Desc)
	go p.Describe(ch)

	for _, name := range []string{
		"covid_reported_count",
	} {
		metric := <-ch
		assert.Contains(t, metric.String(), "\""+name+"\"")
	}
}
