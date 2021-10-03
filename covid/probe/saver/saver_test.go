package saver_test

import (
	"fmt"
	"github.com/clambin/covid19/covid/probe/saver"
	mockCovidStore "github.com/clambin/covid19/covid/store/mocks"
	"github.com/clambin/covid19/models"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestStoreSaver_SaveNewEntries(t *testing.T) {
	db := &mockCovidStore.CovidStore{}
	timeStamp := time.Now()
	db.
		On("GetLatestForCountries", []string{"Belgium", "US"}).
		Return(
			map[string]*models.CountryEntry{
				"Belgium": {Timestamp: timeStamp, Name: "Belgium", Code: "BE", Confirmed: 10, Deaths: 2, Recovered: 1},
				"US":      {Timestamp: timeStamp, Name: "US", Code: "US", Confirmed: 100, Deaths: 20, Recovered: 10},
			},
			nil,
		).
		Once()
	db.
		On("Add", []*models.CountryEntry{{Timestamp: timeStamp.Add(24 * time.Hour), Name: "US", Code: "US", Confirmed: 120, Deaths: 25, Recovered: 10}}).
		Return(nil).
		Once()

	s := saver.StoreSaver{Store: db}

	newEntries, err := s.SaveNewEntries([]*models.CountryEntry{
		{Timestamp: timeStamp.Add(-24 * time.Hour), Name: "Belgium", Code: "BE", Confirmed: 8, Deaths: 1, Recovered: 0},
		{Timestamp: timeStamp.Add(24 * time.Hour), Name: "US", Code: "US", Confirmed: 120, Deaths: 25, Recovered: 10},
	})

	require.NoError(t, err)
	require.Len(t, newEntries, 1)

	mock.AssertExpectationsForObjects(t, db)
}

func TestStoreSaver_SaveNewEntries_Errors(t *testing.T) {
	db := &mockCovidStore.CovidStore{}
	timeStamp := time.Now()
	db.
		On("GetLatestForCountries", []string{"Belgium", "US"}).
		Return(nil, fmt.Errorf("unable to get latest entries for countries")).
		Once()

	s := saver.StoreSaver{Store: db}

	_, err := s.SaveNewEntries([]*models.CountryEntry{
		{Timestamp: timeStamp.Add(-24 * time.Hour), Name: "Belgium", Code: "BE", Confirmed: 8, Deaths: 1, Recovered: 0},
		{Timestamp: timeStamp.Add(24 * time.Hour), Name: "US", Code: "US", Confirmed: 120, Deaths: 25, Recovered: 10},
	})

	require.Error(t, err)

	db.
		On("GetLatestForCountries", []string{"Belgium", "US"}).
		Return(
			map[string]*models.CountryEntry{
				"Belgium": {Timestamp: timeStamp, Name: "Belgium", Code: "BE", Confirmed: 10, Deaths: 2, Recovered: 1},
				"US":      {Timestamp: timeStamp, Name: "US", Code: "US", Confirmed: 100, Deaths: 20, Recovered: 10},
			},
			nil,
		).
		Once()
	db.
		On("Add", []*models.CountryEntry{{Timestamp: timeStamp.Add(24 * time.Hour), Name: "US", Code: "US", Confirmed: 120, Deaths: 25, Recovered: 10}}).
		Return(fmt.Errorf("unable to store records")).
		Once()

	_, err = s.SaveNewEntries([]*models.CountryEntry{
		{Timestamp: timeStamp.Add(-24 * time.Hour), Name: "Belgium", Code: "BE", Confirmed: 8, Deaths: 1, Recovered: 0},
		{Timestamp: timeStamp.Add(24 * time.Hour), Name: "US", Code: "US", Confirmed: 120, Deaths: 25, Recovered: 10},
	})

	require.Error(t, err)

	mock.AssertExpectationsForObjects(t, db)
}
