package countries_test

import (
	"context"
	"errors"
	mockCovidStore "github.com/clambin/covid19/covid/store/mocks"
	"github.com/clambin/covid19/models"
	"github.com/clambin/covid19/simplejsonserver/countries"
	"github.com/clambin/simplejson/v3/common"
	"github.com/clambin/simplejson/v3/query"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestConfirmedByCountry(t *testing.T) {
	dbh := &mockCovidStore.CovidStore{}

	timestamp := time.Date(2022, 1, 26, 0, 0, 0, 0, time.UTC)
	dbh.On("GetAllCountryNames").Return([]string{"A", "B"}, nil)
	dbh.On("GetLatestForCountriesByTime", []string{"A", "B"}, mock.AnythingOfType("time.Time")).Return(map[string]models.CountryEntry{
		"A": {Timestamp: timestamp, Confirmed: 3},
		"B": {Timestamp: timestamp, Confirmed: 10},
	}, nil)
	dbh.On("GetLatestForCountries", []string{"A", "B"}).Return(map[string]models.CountryEntry{
		"A": {Timestamp: timestamp, Confirmed: 3},
		"B": {Timestamp: timestamp, Confirmed: 10},
	}, nil)

	h := countries.ByCountryHandler{
		CovidDB: dbh,
		Mode:    countries.CountryConfirmed,
	}

	ctx := context.Background()

	for _, args := range []query.Args{
		{Args: common.Args{Range: common.Range{To: timestamp}}},
		{},
	} {
		response, err := h.Endpoints().Query(ctx, query.Request{Args: args})
		require.NoError(t, err)
		assert.Equal(t, &query.TableResponse{Columns: []query.Column{
			{Text: "timestamp", Data: query.TimeColumn{timestamp}},
			{Text: "A", Data: query.NumberColumn{3}},
			{Text: "B", Data: query.NumberColumn{10}},
		}}, response)
	}

	mock.AssertExpectationsForObjects(t, dbh)
}

func TestDeathsByCountry(t *testing.T) {
	dbh := &mockCovidStore.CovidStore{}

	timestamp := time.Date(2022, 1, 26, 0, 0, 0, 0, time.UTC)
	dbh.On("GetAllCountryNames").Return([]string{"A", "B"}, nil)
	dbh.On("GetLatestForCountriesByTime", []string{"A", "B"}, mock.AnythingOfType("time.Time")).Return(map[string]models.CountryEntry{
		"A": {Timestamp: timestamp, Deaths: 0},
		"B": {Timestamp: timestamp, Deaths: 1},
	}, nil)
	dbh.On("GetLatestForCountries", []string{"A", "B"}).Return(map[string]models.CountryEntry{
		"A": {Timestamp: timestamp, Deaths: 0},
		"B": {Timestamp: timestamp, Deaths: 1},
	}, nil)

	h := countries.ByCountryHandler{
		CovidDB: dbh,
		Mode:    countries.CountryDeaths,
	}

	ctx := context.Background()
	for _, args := range []query.Args{
		{Args: common.Args{Range: common.Range{To: time.Now()}}},
		{},
	} {
		response, err := h.Endpoints().Query(ctx, query.Request{Args: args})
		require.NoError(t, err)
		assert.Equal(t, &query.TableResponse{Columns: []query.Column{
			{Text: "timestamp", Data: query.TimeColumn{timestamp}},
			{Text: "A", Data: query.NumberColumn{0}},
			{Text: "B", Data: query.NumberColumn{1}},
		}}, response)
	}

	mock.AssertExpectationsForObjects(t, dbh)
}

func TestConfirmedByCountry_Errors(t *testing.T) {
	dbh := &mockCovidStore.CovidStore{}

	h := countries.ByCountryHandler{
		CovidDB: dbh,
		Mode:    countries.CountryConfirmed,
	}

	ctx := context.Background()
	args := query.Args{Args: common.Args{Range: common.Range{To: time.Now()}}}

	dbh.On("GetAllCountryNames").Return(nil, errors.New("db error")).Once()
	_, err := h.Endpoints().Query(ctx, query.Request{Args: args})
	assert.Error(t, err)

	dbh.On("GetAllCountryNames").Return([]string{"A", "B"}, nil).Once()
	dbh.On("GetLatestForCountriesByTime", []string{"A", "B"}, mock.AnythingOfType("time.Time")).Return(nil, errors.New("db error")).Once()

	_, err = h.Endpoints().Query(ctx, query.Request{Args: args})
	assert.Error(t, err)

	mock.AssertExpectationsForObjects(t, dbh)
}
