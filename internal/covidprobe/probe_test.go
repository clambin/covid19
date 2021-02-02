package covidprobe_test

import (
	"covid19/internal/configuration"
	"covid19/internal/coviddb"
	mockdb "covid19/internal/coviddb/mock"
	"covid19/internal/covidprobe"
	"covid19/internal/covidprobe/mockapi"
	"github.com/clambin/gotools/metrics"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

var (
	lastUpdate = time.Date(2020, time.December, 3, 5, 28, 22, 0, time.UTC)

	seedDB = []coviddb.CountryEntry{
		{
			Timestamp: time.Date(2020, time.November, 1, 0, 0, 0, 0, time.UTC),
			Code:      "BE",
			Name:      "Belgium",
			Confirmed: 1,
			Recovered: 0,
			Deaths:    0},
		{
			Timestamp: time.Date(2020, time.November, 2, 0, 0, 0, 0, time.UTC),
			Code:      "BE",
			Name:      "Belgium",
			Confirmed: 3,
			Recovered: 0,
			Deaths:    0},
		{
			Timestamp: time.Date(2020, time.November, 2, 0, 0, 0, 0, time.UTC),
			Code:      "US",
			Name:      "US",
			Confirmed: 3,
			Recovered: 1,
			Deaths:    0},
		{
			Timestamp: time.Date(2020, time.November, 4, 0, 0, 0, 0, time.UTC),
			Code:      "BE",
			Name:      "Belgium",
			Confirmed: 10,
			Recovered: 5,
			Deaths:    1,
		},
	}
)

func TestProbe(t *testing.T) {
	dbh := mockdb.Create(seedDB)

	cfg := configuration.MonitorConfiguration{
		Enabled: true,
		Notifications: configuration.NotificationsConfiguration{
			Enabled: false,
			URL:     "",
			Countries: []string{
				"Belgium",
			},
		},
	}
	p := covidprobe.NewProbe(&cfg, dbh)
	p.APIClient = mockapi.New(map[string]covidprobe.CountryStats{
		"Belgium":     {LastUpdate: lastUpdate, Confirmed: 4, Recovered: 2, Deaths: 1},
		"US":          {LastUpdate: lastUpdate, Confirmed: 20, Recovered: 15, Deaths: 5},
		"NotACountry": {LastUpdate: lastUpdate, Confirmed: 0, Recovered: 0, Deaths: 0},
	})

	err := p.Run()

	assert.Nil(t, err)

	latest, err := dbh.ListLatestByCountry()

	assert.Nil(t, err)
	assert.Len(t, latest, 2)
	assert.True(t, latest["Belgium"].Equal(lastUpdate))
	assert.True(t, latest["US"].Equal(lastUpdate))

	var (
		value float64
	)
	value, err = metrics.LoadValue("covid_reported_count", "Belgium")
	assert.Nil(t, err)
	assert.Equal(t, 1.0, value)
	value, err = metrics.LoadValue("covid_reported_count", "US")
	assert.Nil(t, err)
	assert.Equal(t, 1.0, value)
}

func TestProbeCaches(t *testing.T) {
	dbh := mockdb.Create(seedDB)
	cfg := configuration.MonitorConfiguration{
		Enabled: true,
		Notifications: configuration.NotificationsConfiguration{
			Enabled: true,
			URL:     "",
			Countries: []string{
				"Belgium", "Sokovia", "France",
			},
		},
	}
	p := covidprobe.NewProbe(&cfg, dbh)

	// LastUpdates should not contain the latest timestamp for each country in the DB
	assert.Len(t, p.LatestUpdates, 2)
	timestamp, ok := p.LatestUpdates["Belgium"]
	assert.True(t, ok)
	assert.Equal(t, time.Date(2020, time.November, 4, 0, 0, 0, 0, time.UTC), timestamp)
	timestamp, ok = p.LatestUpdates["US"]
	assert.True(t, ok)
	assert.Equal(t, time.Date(2020, time.November, 2, 0, 0, 0, 0, time.UTC), timestamp)

	// NotifyCache should contain the latest entry for each (valid) country we want to send notifications for
	assert.Len(t, p.NotifyCache, 2)
	record, ok := p.NotifyCache["Belgium"]
	assert.True(t, ok)
	assert.Equal(t, int64(10), record.Confirmed)
	record, ok = p.NotifyCache["France"]
	assert.True(t, ok)
	assert.Equal(t, int64(0), record.Confirmed)

	p.APIClient = mockapi.New(map[string]covidprobe.CountryStats{
		"Belgium":     {LastUpdate: lastUpdate, Confirmed: 40, Recovered: 10, Deaths: 1},
		"US":          {LastUpdate: lastUpdate, Confirmed: 20, Recovered: 15, Deaths: 5},
		"NotACountry": {LastUpdate: lastUpdate, Confirmed: 0, Recovered: 0, Deaths: 0},
	})

	err := p.Run()

	assert.NotNil(t, err)
	// FIXME: test is dependent on shoutrrr implementation. needs a more generic test
	assert.Equal(t, "error sending message: no senders", err.Error())

	// Check that LatestUpdates cache was updated
	assert.Len(t, p.LatestUpdates, 2)
	timestamp, ok = p.LatestUpdates["Belgium"]
	assert.True(t, ok)
	assert.Equal(t, lastUpdate, timestamp)
	timestamp, ok = p.LatestUpdates["US"]
	assert.True(t, ok)
	assert.Equal(t, lastUpdate, timestamp)

	// Check that the NotifyCache was updated
	// NotifyCache should contain the latest entry for each (valid) country we want to send notifications for
	assert.Len(t, p.NotifyCache, 2)
	record, ok = p.NotifyCache["Belgium"]
	assert.True(t, ok)
	assert.Equal(t, "Belgium", record.Name)
	assert.Equal(t, int64(40), record.Confirmed)
	record, ok = p.NotifyCache["France"]
	assert.True(t, ok)
	assert.Equal(t, int64(0), record.Confirmed)
}
