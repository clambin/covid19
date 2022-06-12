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

var testData = []models.CountryEntry{
	{
		Timestamp: time.Date(2020, time.November, 1, 0, 0, 0, 0, time.UTC),
		Confirmed: 1,
		Recovered: 0,
		Deaths:    0},
	{
		Timestamp: time.Date(2020, time.November, 2, 0, 0, 0, 0, time.UTC),
		Confirmed: 3,
		Recovered: 0,
		Deaths:    0},
	{
		Timestamp: time.Date(2020, time.November, 3, 0, 0, 0, 0, time.UTC),
		Confirmed: 3,
		Recovered: 1,
		Deaths:    0},
	{
		Timestamp: time.Date(2020, time.November, 4, 0, 0, 0, 0, time.UTC),
		Confirmed: 10,
		Recovered: 5,
		Deaths:    1},
}

func TestCovidCache_Totals(t *testing.T) {
	db := &mocks.CovidStore{}
	c := &cache.Cache{DB: db, Retention: 20 * time.Minute}

	// Set up expectations
	db.On("GetTotalsPerDay").Return(testData, nil).Once()

	response, err := c.GetTotals(time.Now())
	require.NoError(t, err)
	require.Len(t, response.GetColumns(), 3)

	assert.Equal(t, []time.Time{
		time.Date(2020, time.November, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2020, time.November, 2, 0, 0, 0, 0, time.UTC),
		time.Date(2020, time.November, 3, 0, 0, 0, 0, time.UTC),
		time.Date(2020, time.November, 4, 0, 0, 0, 0, time.UTC),
	}, response.GetTimestamps())

	values, ok := response.GetFloatValues("confirmed")
	require.True(t, ok)
	assert.Equal(t, []float64{1, 3, 3, 10}, values)

	values, ok = response.GetFloatValues("deaths")
	require.True(t, ok)
	assert.Equal(t, []float64{0, 0, 0, 1}, values)

	mock.AssertExpectationsForObjects(t, db)
}

func TestGetTotalDeltas(t *testing.T) {
	db := &mocks.CovidStore{}
	c := &cache.Cache{DB: db, Retention: 20 * time.Minute}

	// Set up expectations
	db.On("GetTotalsPerDay").Return(testData, nil).Once()

	response, err := c.GetDeltas(time.Now())
	require.NoError(t, err)
	require.Len(t, response.GetColumns(), 3)

	assert.Equal(t, []time.Time{
		time.Date(2020, time.November, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2020, time.November, 2, 0, 0, 0, 0, time.UTC),
		time.Date(2020, time.November, 3, 0, 0, 0, 0, time.UTC),
		time.Date(2020, time.November, 4, 0, 0, 0, 0, time.UTC),
	}, response.GetTimestamps())

	values, ok := response.GetFloatValues("confirmed")
	require.True(t, ok)
	assert.Equal(t, []float64{1, 2, 0, 7}, values)

	values, ok = response.GetFloatValues("deaths")
	require.True(t, ok)
	assert.Equal(t, []float64{0, 0, 0, 1}, values)

	mock.AssertExpectationsForObjects(t, db)
}

func TestCache_Errors(t *testing.T) {
	db := &mocks.CovidStore{}
	c := &cache.Cache{DB: db, Retention: 20 * time.Minute}

	db.On("GetTotalsPerDay").Return(nil, fmt.Errorf("database error"))

	_, err := c.GetTotals(time.Now())
	require.Error(t, err)

	_, err = c.GetDeltas(time.Now())
	require.Error(t, err)
}
