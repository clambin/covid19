package covidprobe_test

import (
	"covid19/internal/coviddb"
	mockdb "covid19/internal/coviddb/mock"
	"covid19/internal/covidprobe"
	"covid19/internal/covidprobe/mockapi"
	"covid19/internal/reporters"

	"github.com/clambin/gotools/metrics"
	"github.com/stretchr/testify/assert"

	"testing"
	"time"
)

var lastUpdate = time.Date(2020, time.December, 3, 5, 28, 22, 0, time.UTC)

func TestProbe(t *testing.T) {
	dbh := mockdb.Create([]coviddb.CountryEntry{
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
	})

	r := reporters.Create()
	r.Add(reporters.NewCountriesReporter("localhost:8080"))

	p := covidprobe.NewProbe("", dbh, r)
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