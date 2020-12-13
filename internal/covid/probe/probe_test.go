package probe_test

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"

	"covid19/internal/covid/apiclient"
	mockapi "covid19/internal/covid/apiclient/mock"
	"covid19/internal/covid/db"
	mockdb "covid19/internal/covid/db/mock"
	"covid19/internal/covid/probe"
)

var lastUpdate = time.Date(2020, time.December, 3, 5, 28, 22, 0, time.UTC)

func TestProbe(t *testing.T) {
	dbh := mockdb.Create([]db.CountryEntry{
		{
			Timestamp: time.Date(2020, time.November, 1, 0, 0, 0, 0, time.UTC),
			Code:      "A",
			Name:      "A",
			Confirmed: 1,
			Recovered: 0,
			Deaths:    0},
		{
			Timestamp: time.Date(2020, time.November, 2, 0, 0, 0, 0, time.UTC),
			Code:      "B",
			Name:      "B",
			Confirmed: 3,
			Recovered: 0,
			Deaths:    0},
		{
			Timestamp: time.Date(2020, time.November, 2, 0, 0, 0, 0, time.UTC),
			Code:      "A",
			Name:      "A",
			Confirmed: 3,
			Recovered: 1,
			Deaths:    0},
		{
			Timestamp: time.Date(2020, time.November, 4, 0, 0, 0, 0, time.UTC),
			Code:      "B",
			Name:      "B",
			Confirmed: 10,
			Recovered: 5,
			Deaths:    1,
		},
	})
	apiClient := mockapi.New(map[string]apiclient.CountryStats{
		"A": {LastUpdate: lastUpdate, Confirmed: 4, Recovered: 2, Deaths: 1},
		"B": {LastUpdate: lastUpdate, Confirmed: 20, Recovered: 15, Deaths: 5},
	})

	p := probe.NewProbe(apiClient, dbh, nil)

	err := p.Run()

	assert.Nil(t, err)

	latest, err := dbh.ListLatestByCountry()

	assert.Nil(t, err)
	assert.Len(t, latest, 2)
	assert.True(t, latest["A"].Equal(lastUpdate))
	assert.True(t, latest["B"].Equal(lastUpdate))
}
