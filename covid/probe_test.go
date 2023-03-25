package covid_test

import (
	"context"
	"github.com/clambin/covid19/configuration"
	"github.com/clambin/covid19/covid"
	mockFetcher "github.com/clambin/covid19/covid/fetcher/mocks"
	mockRouter "github.com/clambin/covid19/covid/shoutrrr/mocks"
	covid2 "github.com/clambin/covid19/internal/testtools/db/covid"
	"github.com/clambin/covid19/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestCovid19Probe_Update(t *testing.T) {
	cfg := configuration.MonitorConfiguration{
		RapidAPIKey: "1234",
		Notifications: configuration.NotificationConfiguration{
			Enabled:   true,
			URL:       "slack://T0000000000/B0000000000/I00000000000000000000000",
			Countries: []string{"US"},
		},
	}

	timeStamp := time.Now()
	fdb := covid2.FakeStore{Records: []models.CountryEntry{
		{Timestamp: timeStamp, Name: "Belgium", Code: "BE", Confirmed: 10, Deaths: 2, Recovered: 1},
		{Timestamp: timeStamp, Name: "US", Code: "US", Confirmed: 100, Deaths: 20, Recovered: 10},
	}}
	f := mockFetcher.NewFetcher(t)
	s := mockRouter.NewSender(t)

	p := covid.New(&cfg, &fdb)
	p.Fetcher = f
	p.StoreSaver.Store = &fdb
	p.Notifier.Sender = s

	countryStats := []models.CountryEntry{
		{Timestamp: timeStamp.Add(-24 * time.Hour), Name: "Belgium", Code: "BE", Confirmed: 8, Deaths: 1, Recovered: 1},
		{Timestamp: timeStamp.Add(24 * time.Hour), Name: "US", Code: "US", Confirmed: 120, Deaths: 25, Recovered: 15},
		{Timestamp: timeStamp.Add(24 * time.Hour), Name: "notacountry", Code: "???", Confirmed: 120, Deaths: 25, Recovered: 15},
	}

	f.
		On("Fetch", mock.AnythingOfType("*context.emptyCtx")).
		Return(countryStats, nil).
		Once()
	s.
		On("Send", "New data for US", "Confirmed: 20, deaths: 5").
		Return(nil).
		Once()

	_, err := p.Update(context.Background())
	require.NoError(t, err)

	latest, err := fdb.GetLatestForCountries(time.Time{})
	require.NoError(t, err)
	assert.Equal(t, map[string]models.CountryEntry{
		"Belgium": {Timestamp: timeStamp, Code: "BE", Name: "Belgium", Confirmed: 10, Recovered: 1, Deaths: 2},
		"US":      {Timestamp: timeStamp.Add(24 * time.Hour), Code: "US", Name: "US", Confirmed: 120, Recovered: 15, Deaths: 25},
	}, latest)
}

/*
func TestCovid19Probe_Update_Errors(t *testing.T) {
	f := mockFetcher.NewFetcher(t)
	s := mockSaver.NewSaver(t)
	r := mockRouter.NewRouter(t)
	db := mockCovidStore.NewCovidStore(t)
	db.On("GetLatestForCountries", time.Time{}).Return(map[string]models.CountryEntry{
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


*/
