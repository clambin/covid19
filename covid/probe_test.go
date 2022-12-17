package covid_test

import (
	"context"
	"fmt"
	"github.com/clambin/covid19/configuration"
	"github.com/clambin/covid19/covid"
	mockFetcher "github.com/clambin/covid19/covid/fetcher/mocks"
	"github.com/clambin/covid19/covid/notifier"
	mockRouter "github.com/clambin/covid19/covid/notifier/mocks"
	mockSaver "github.com/clambin/covid19/covid/saver/mocks"
	mockCovidStore "github.com/clambin/covid19/db/mocks"
	"github.com/clambin/covid19/models"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestCovid19Probe_Update(t *testing.T) {
	cfg := configuration.MonitorConfiguration{
		RapidAPIKey: "1234",
	}
	db := mockCovidStore.NewCovidStore(t)
	timeStamp := time.Now()
	db.
		On("GetLatestForCountries").
		Return(
			map[string]models.CountryEntry{
				"Belgium": {Timestamp: timeStamp, Name: "Belgium", Code: "BE", Confirmed: 10, Deaths: 2, Recovered: 1},
				"US":      {Timestamp: timeStamp, Name: "US", Code: "US", Confirmed: 100, Deaths: 20, Recovered: 10},
			},
			nil,
		)
	p := covid.New(&cfg, db)
	reg := prometheus.NewRegistry()
	reg.MustRegister(p)

	f := mockFetcher.NewFetcher(t)
	r := mockRouter.NewRouter(t)
	n, _ := notifier.NewNotifier(r, []string{"Belgium", "US"}, db)
	s := mockSaver.NewSaver(t)

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
	r.
		On("Send", "New probe data for US", "Confirmed: 20, deaths: 5, recovered: 5").
		Return(nil).
		Once()

	_, err := p.Update(context.Background())
	require.NoError(t, err)

	metrics, err := reg.Gather()
	require.NoError(t, err)
	for _, metric := range metrics {
		assert.Equal(t, "covid_reported_count", metric.GetName())
		for _, m := range metric.GetMetric() {
			labels := m.GetLabel()
			require.Len(t, labels, 1)
			assert.Equal(t, "country", labels[0].GetName())
			var target float64
			switch labels[0].GetValue() {
			case "US":
				target = 1.0
			}
			assert.Equal(t, target, m.Counter.GetValue(), labels[0].GetValue())
		}
	}
}

func TestCovid19Probe_Update_Errors(t *testing.T) {
	f := mockFetcher.NewFetcher(t)
	s := mockSaver.NewSaver(t)
	r := mockRouter.NewRouter(t)
	db := mockCovidStore.NewCovidStore(t)
	db.On("GetLatestForCountries").Return(map[string]models.CountryEntry{
		"US":      {},
		"Belgium": {},
	}, nil)
	n, _ := notifier.NewNotifier(r, []string{"Belgium", "US"}, db)

	p := covid.Probe{
		Fetcher:  f,
		Saver:    s,
		Notifier: n,
	}

	f.
		On("GetCountryStats", mock.AnythingOfType("*context.emptyCtx")).
		Return(nil, fmt.Errorf("unable to get new country stats")).
		Once()

	_, err := p.Update(context.Background())
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

	_, err = p.Update(context.Background())
	require.Error(t, err)

	f.
		On("GetCountryStats", mock.AnythingOfType("*context.emptyCtx")).
		Return(countryStats, nil).
		Once()
	s.
		On("SaveNewEntries", countryStats).
		Return(newCountryStats, nil).
		Once()
	r.
		On("Send", "New probe data for US", "Confirmed: 120, deaths: 25, recovered: 15").
		Return(fmt.Errorf("unable to send notifications")).
		Once()

	_, err = p.Update(context.Background())
	require.NoError(t, err)
}
