package cache_test

import (
	"fmt"
	"github.com/clambin/covid19/cache"
	"github.com/clambin/covid19/covid/store/mocks"
	"github.com/clambin/covid19/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

var (
	testData = []*models.CountryEntry{
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

	totalCases = []cache.Entry{
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

	deltaCases = []cache.Entry{
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
	db := &mocks.CovidStore{}
	c := &cache.Cache{DB: db, Retention: 20 * time.Minute}

	// Set up expectations
	db.On("GetAll").Return(testData, nil).Once()

	response, err := c.GetTotals(time.Now())
	require.NoError(t, err)
	require.Len(t, response, 3)

	for index, totalCase := range totalCases {
		assert.Equal(t, totalCase.Timestamp, response[index].Timestamp)
		assert.Equal(t, totalCase.Confirmed, response[index].Confirmed)
		assert.Equal(t, totalCase.Recovered, response[index].Recovered)
		assert.Equal(t, totalCase.Deaths, response[index].Deaths)
		assert.Equal(t, totalCase.Active, response[index].Active)
	}

	response, err = c.GetDeltas(time.Now())
	require.NoError(t, err)
	require.Len(t, response, 3)

	for index, deltaCase := range deltaCases {
		assert.Equal(t, deltaCase.Timestamp, response[index].Timestamp)
		assert.Equal(t, deltaCase.Confirmed, response[index].Confirmed)
		assert.Equal(t, deltaCase.Recovered, response[index].Recovered)
		assert.Equal(t, deltaCase.Deaths, response[index].Deaths)
		assert.Equal(t, deltaCase.Active, response[index].Active)
	}

	response, err = c.GetTotals(time.Date(2020, time.November, 3, 0, 0, 0, 0, time.UTC))
	require.NoError(t, err)
	require.Len(t, response, 2)

	for index, total := range response {
		assert.Equal(t, totalCases[index].Timestamp, total.Timestamp)
		assert.Equal(t, totalCases[index].Confirmed, total.Confirmed)
		assert.Equal(t, totalCases[index].Recovered, total.Recovered)
		assert.Equal(t, totalCases[index].Deaths, total.Deaths)
		assert.Equal(t, totalCases[index].Active, total.Active)
	}

	mock.AssertExpectationsForObjects(t, db)
}

func TestCache_Errors(t *testing.T) {
	db := &mocks.CovidStore{}
	c := &cache.Cache{DB: db, Retention: 20 * time.Minute}

	db.On("GetAll").Return(nil, fmt.Errorf("database error"))

	_, err := c.GetTotals(time.Now())
	require.Error(t, err)

	_, err = c.GetDeltas(time.Now())
	require.Error(t, err)
}
