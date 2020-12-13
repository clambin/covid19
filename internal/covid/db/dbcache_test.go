package db_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"covid19/internal/covid/db"
	"covid19/internal/covid/db/mock"
)

func TestDBCache_List(t *testing.T) {
	covidDB := mock.Create([]db.CountryEntry{
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
			Deaths:    1},
	})

	cache := db.NewCache(covidDB, 5*time.Minute)

	assert.NotNil(t, cache)

	entries, err := cache.List(time.Now())
	assert.Nil(t, err)
	assert.Len(t, entries, 4)

	entries, err = cache.List(time.Date(2020, time.November, 2, 0, 0, 0, 0, time.UTC))
	assert.Nil(t, err)
	assert.Len(t, entries, 1)
	assert.Equal(t, "BE", entries[0].Code)
	assert.Equal(t, int64(1), entries[0].Confirmed)
	assert.Equal(t, int64(0), entries[0].Deaths)
	assert.Equal(t, int64(0), entries[0].Recovered)

	// Insert a record before 2020-11-02, list and again and validate that it's returned
	// (i.e. we're still getting the cached version)

	_ = covidDB.Add([]db.CountryEntry{
		{
			Timestamp: time.Date(2020, time.November, 1, 0, 0, 0, 0, time.UTC),
			Code:      "US",
			Name:      "United States",
			Confirmed: 3,
			Recovered: 0,
			Deaths:    0,
		},
	})
	entries, err = cache.List(time.Date(2020, time.November, 2, 0, 0, 0, 0, time.UTC))
	assert.Nil(t, err)
	assert.Len(t, entries, 1)
}
