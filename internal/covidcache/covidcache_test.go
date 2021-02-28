package covidcache_test

import (
	"covid19/internal/covidcache"
	"covid19/internal/coviddb"
	mockdb "covid19/internal/coviddb/mock"
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
	db := covidcache.Cache{
		DB: mockdb.Create(testData),
	}

	err := db.Update()
	assert.Nil(t, err)

	totals := db.GetTotals(time.Now())
	assert.Len(t, totals, 3)

	for index, totalCase := range totalCases {
		assert.Equal(t, totalCase.Timestamp, totals[index].Timestamp)
		assert.Equal(t, totalCase.Confirmed, totals[index].Confirmed)
		assert.Equal(t, totalCase.Recovered, totals[index].Recovered)
		assert.Equal(t, totalCase.Deaths, totals[index].Deaths)
		assert.Equal(t, totalCase.Active, totals[index].Active)
	}

	deltas := db.GetDeltas(time.Now())
	assert.Len(t, deltas, 3)

	for index, deltaCase := range deltaCases {
		assert.Equal(t, deltaCase.Timestamp, deltas[index].Timestamp)
		assert.Equal(t, deltaCase.Confirmed, deltas[index].Confirmed)
		assert.Equal(t, deltaCase.Recovered, deltas[index].Recovered)
		assert.Equal(t, deltaCase.Deaths, deltas[index].Deaths)
		assert.Equal(t, deltaCase.Active, deltas[index].Active)
	}

	totals = db.GetTotals(time.Date(2020, time.November, 3, 0, 0, 0, 0, time.UTC))
	assert.Len(t, totals, 2)

	for index, total := range totals {
		assert.Equal(t, totalCases[index].Timestamp, total.Timestamp)
		assert.Equal(t, totalCases[index].Confirmed, total.Confirmed)
		assert.Equal(t, totalCases[index].Recovered, total.Recovered)
		assert.Equal(t, totalCases[index].Deaths, total.Deaths)
		assert.Equal(t, totalCases[index].Active, total.Active)
	}
}
