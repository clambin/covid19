package covidcache_test

import (
	"github.com/clambin/covid19/internal/covidcache"
	"github.com/clambin/covid19/internal/coviddb"
	mockdb "github.com/clambin/covid19/internal/coviddb/mock"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

var (
	testData = []coviddb.CountryEntry{
		{
			Timestamp: time.Date(2020, time.November, 1, 0, 0, 0, 0, time.UTC),
			Code:      "BE",
			Name:      "Belgium",
			Confirmed: 1,
			Recovered: 0,
			Deaths:    0},
		{
			Timestamp: time.Date(2020, time.November, 2, 0, 0, 0, 0, time.UTC),
			Code:      "US",
			Name:      "United States",
			Confirmed: 3,
			Recovered: 0,
			Deaths:    0},
		{
			Timestamp: time.Date(2020, time.November, 2, 0, 0, 0, 0, time.UTC),
			Code:      "BE",
			Name:      "Belgium",
			Confirmed: 3,
			Recovered: 1,
			Deaths:    0},
		{
			Timestamp: time.Date(2020, time.November, 4, 0, 0, 0, 0, time.UTC),
			Code:      "US",
			Name:      "United States",
			Confirmed: 10,
			Recovered: 5,
			Deaths:    1}}

	totalCases = []covidcache.CacheEntry{
		{
			Timestamp: time.Date(2020, time.November, 1, 0, 0, 0, 0, time.UTC),
			Confirmed: 1,
			Recovered: 0,
			Deaths:    0,
			Active:    1,
		},
		{
			Timestamp: time.Date(2020, time.November, 2, 0, 0, 0, 0, time.UTC),
			Confirmed: 6,
			Recovered: 1,
			Deaths:    0,
			Active:    5,
		},
		{
			Timestamp: time.Date(2020, time.November, 4, 0, 0, 0, 0, time.UTC),
			Confirmed: 13,
			Recovered: 6,
			Deaths:    1,
			Active:    6,
		},
	}

	deltaCases = []covidcache.CacheEntry{
		{
			Timestamp: time.Date(2020, time.November, 1, 0, 0, 0, 0, time.UTC),
			Confirmed: 1,
			Recovered: 0,
			Deaths:    0,
			Active:    1,
		},
		{
			Timestamp: time.Date(2020, time.November, 2, 0, 0, 0, 0, time.UTC),
			Confirmed: 5,
			Recovered: 1,
			Deaths:    0,
			Active:    4,
		},
		{
			Timestamp: time.Date(2020, time.November, 4, 0, 0, 0, 0, time.UTC),
			Confirmed: 7,
			Recovered: 5,
			Deaths:    1,
			Active:    1,
		},
	}
)

func TestCovidCache(t *testing.T) {
	cache := covidcache.New(mockdb.Create(testData))
	responseChannel := make(chan covidcache.Response)
	go cache.Run()

	req := covidcache.Request{
		Response: responseChannel,
		End:      time.Now(),
		Delta:    false,
	}

	cache.Refresh <- true
	cache.Request <- req
	response := <-responseChannel
	assert.Len(t, response.Series, 3)

	for index, totalCase := range totalCases {
		assert.Equal(t, totalCase.Timestamp, response.Series[index].Timestamp)
		assert.Equal(t, totalCase.Confirmed, response.Series[index].Confirmed)
		assert.Equal(t, totalCase.Recovered, response.Series[index].Recovered)
		assert.Equal(t, totalCase.Deaths, response.Series[index].Deaths)
		assert.Equal(t, totalCase.Active, response.Series[index].Active)
	}

	req.Delta = true
	cache.Request <- req
	response = <-responseChannel
	assert.Len(t, response.Series, 3)

	for index, deltaCase := range deltaCases {
		assert.Equal(t, deltaCase.Timestamp, response.Series[index].Timestamp)
		assert.Equal(t, deltaCase.Confirmed, response.Series[index].Confirmed)
		assert.Equal(t, deltaCase.Recovered, response.Series[index].Recovered)
		assert.Equal(t, deltaCase.Deaths, response.Series[index].Deaths)
		assert.Equal(t, deltaCase.Active, response.Series[index].Active)
	}

	cache.Request <- covidcache.Request{
		Response: responseChannel,
		End:      time.Date(2020, time.November, 3, 0, 0, 0, 0, time.UTC),
		Delta:    false,
	}
	response = <-responseChannel
	assert.Len(t, response.Series, 2)

	for index, total := range response.Series {
		assert.Equal(t, totalCases[index].Timestamp, total.Timestamp)
		assert.Equal(t, totalCases[index].Confirmed, total.Confirmed)
		assert.Equal(t, totalCases[index].Recovered, total.Recovered)
		assert.Equal(t, totalCases[index].Deaths, total.Deaths)
		assert.Equal(t, totalCases[index].Active, total.Active)
	}
}
