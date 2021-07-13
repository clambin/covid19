package covidprobe_test

import (
	"context"
	"github.com/clambin/covid19/configuration"
	"github.com/clambin/covid19/covidcache"
	"github.com/clambin/covid19/coviddb"
	"github.com/clambin/covid19/coviddb/mock"
	"github.com/clambin/covid19/covidprobe"
	"github.com/clambin/covid19/covidprobe/mockapi"
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
	ctx, cancel := context.WithCancel(context.Background())

	dbh := mock.Create(seedDB)
	cache := covidcache.New(dbh)
	go cache.Run(ctx)

	cfg := configuration.MonitorConfiguration{
		Enabled: true,
		Notifications: configuration.NotificationConfiguration{
			Enabled: true,
			URL:     configuration.ValueOrEnvVar{Value: ""},
			Countries: []string{
				"Belgium", "Sokovia", "France",
			},
		},
	}

	p, err := covidprobe.NewProbe(&cfg, dbh, cache)
	assert.NoError(t, err)

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

	go func() {
		_ = p.Run(ctx, 24*time.Hour)
	}()

	// Check that the latest values were added to the DB
	var latest map[string]time.Time
	assert.Eventually(t, func() bool {
		latest, err = dbh.ListLatestByCountry()
		return err == nil && len(latest) == 2
	}, 500*time.Millisecond, 10*time.Millisecond)

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

	cancel()
	ctx, cancel = context.WithCancel(context.Background())
	defer cancel()

	go func() {
		_ = p.Run(ctx, 24*time.Hour)
	}()

	// Prometheus metrics should now be zero
	assert.Eventually(t, func() bool {
		value, err = metrics.LoadValue("covid_reported_count", "Belgium")
		if err != nil || value != 0.0 {
			return false
		}
		value, err = metrics.LoadValue("covid_reported_count", "US")
		if err != nil || value != 0.0 {
			return false
		}
		return true
	}, 500*time.Millisecond, 10*time.Millisecond)
}
