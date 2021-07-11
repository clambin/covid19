package covidcache_test

import (
	"github.com/clambin/covid19/covidcache"
	"github.com/clambin/covid19/coviddb"
	"github.com/clambin/covid19/coviddb/mock"
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
	cache := covidcache.New(mock.Create(testData))
	go cache.Run()

	cache.Refresh()
	response := cache.GetTotals(time.Now())
	assert.Len(t, response, 3)

	for index, totalCase := range totalCases {
		assert.Equal(t, totalCase.Timestamp, response[index].Timestamp)
		assert.Equal(t, totalCase.Confirmed, response[index].Confirmed)
		assert.Equal(t, totalCase.Recovered, response[index].Recovered)
		assert.Equal(t, totalCase.Deaths, response[index].Deaths)
		assert.Equal(t, totalCase.Active, response[index].Active)
	}

	response = cache.GetDeltas(time.Now())
	assert.Len(t, response, 3)

	for index, deltaCase := range deltaCases {
		assert.Equal(t, deltaCase.Timestamp, response[index].Timestamp)
		assert.Equal(t, deltaCase.Confirmed, response[index].Confirmed)
		assert.Equal(t, deltaCase.Recovered, response[index].Recovered)
		assert.Equal(t, deltaCase.Deaths, response[index].Deaths)
		assert.Equal(t, deltaCase.Active, response[index].Active)
	}

	response = cache.GetTotals(time.Date(2020, time.November, 3, 0, 0, 0, 0, time.UTC))
	assert.Len(t, response, 2)

	for index, total := range response {
		assert.Equal(t, totalCases[index].Timestamp, total.Timestamp)
		assert.Equal(t, totalCases[index].Confirmed, total.Confirmed)
		assert.Equal(t, totalCases[index].Recovered, total.Recovered)
		assert.Equal(t, totalCases[index].Deaths, total.Deaths)
		assert.Equal(t, totalCases[index].Active, total.Active)
	}
}
