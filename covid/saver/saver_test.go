package saver_test

import (
	"github.com/clambin/covid19/covid/saver"
	"github.com/clambin/covid19/internal/testtools/db/covid"
	"github.com/clambin/covid19/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestStoreSaver_SaveNewEntries(t *testing.T) {
	timeStamp := time.Now()
	f := covid.FakeStore{Records: []models.CountryEntry{
		{Timestamp: timeStamp, Name: "Belgium", Code: "BE", Confirmed: 10, Deaths: 2, Recovered: 1},
		{Timestamp: timeStamp, Name: "US", Code: "US", Confirmed: 100, Deaths: 20, Recovered: 10},
	}}
	s := saver.StoreSaver{Store: &f}

	newEntries, err := s.SaveNewEntries([]models.CountryEntry{
		{Timestamp: timeStamp.Add(-24 * time.Hour), Name: "Belgium", Code: "BE", Confirmed: 8, Deaths: 1, Recovered: 0},
		{Timestamp: timeStamp.Add(24 * time.Hour), Name: "US", Code: "US", Confirmed: 120, Deaths: 25, Recovered: 10},
	})

	require.NoError(t, err)
	require.Len(t, newEntries, 1)

	n, err := s.Store.GetLatestForCountries(time.Time{})
	require.NoError(t, err)
	assert.Equal(t, map[string]models.CountryEntry{
		"Belgium": {Timestamp: timeStamp, Code: "BE", Name: "Belgium", Confirmed: 10, Recovered: 1, Deaths: 2},
		"US":      {Timestamp: timeStamp.Add(24 * time.Hour), Code: "US", Name: "US", Confirmed: 120, Recovered: 10, Deaths: 25},
	}, n)
}

func TestStoreSaver_SaveNewEntries_Errors(t *testing.T) {
	timeStamp := time.Now()
	f := covid.FakeStore{Fail: true, Records: []models.CountryEntry{
		{Timestamp: timeStamp, Name: "Belgium", Code: "BE", Confirmed: 10, Deaths: 2, Recovered: 1},
		{Timestamp: timeStamp, Name: "US", Code: "US", Confirmed: 100, Deaths: 20, Recovered: 10},
	}}
	s := saver.StoreSaver{Store: &f}

	_, err := s.SaveNewEntries([]models.CountryEntry{
		{Timestamp: timeStamp.Add(-24 * time.Hour), Name: "Belgium", Code: "BE", Confirmed: 8, Deaths: 1, Recovered: 0},
		{Timestamp: timeStamp.Add(24 * time.Hour), Name: "US", Code: "US", Confirmed: 120, Deaths: 25, Recovered: 10},
	})
	require.Error(t, err)
}

var _ saver.CovidAdderGetter = &covid.FakeStore{}
