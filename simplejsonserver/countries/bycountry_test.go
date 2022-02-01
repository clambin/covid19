package countries_test

import (
	"context"
	"errors"
	mockCovidStore "github.com/clambin/covid19/covid/store/mocks"
	"github.com/clambin/covid19/models"
	"github.com/clambin/covid19/simplejsonserver/countries"
	"github.com/clambin/simplejson/v2/common"
	"github.com/clambin/simplejson/v2/query"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestConfirmedByCountry(t *testing.T) {
	dbh := &mockCovidStore.CovidStore{}
	dbh.On("GetAllCountryNames").Return([]string{"A", "B"}, nil)
	dbh.On("GetLatestForCountriesByTime", []string{"A", "B"}, mock.AnythingOfType("time.Time")).Return(map[string]models.CountryEntry{
		"A": {Timestamp: time.Now(), Confirmed: 3},
		"B": {Timestamp: time.Now(), Confirmed: 10},
	}, nil)
	dbh.On("GetLatestForCountries", []string{"A", "B"}).Return(map[string]models.CountryEntry{
		"A": {Timestamp: time.Now(), Confirmed: 3},
		"B": {Timestamp: time.Now(), Confirmed: 10},
	}, nil)

	h := countries.ByCountryHandler{
		CovidDB: dbh,
		Mode:    countries.CountryConfirmed,
	}

	ctx := context.Background()

	for _, args := range []query.Args{
		{Args: common.Args{Range: common.Range{To: time.Now()}}},
		{},
	} {

		response, err := h.Endpoints().TableQuery(ctx, args)
		require.NoError(t, err)
		require.Len(t, response.Columns, 3)
		assert.Equal(t, "timestamp", response.Columns[0].Text)
		assert.Len(t, response.Columns[0].Data, 1)
		assert.Equal(t, query.Column{
			Text: "A",
			Data: query.NumberColumn{3.0},
		}, response.Columns[1])
		assert.Equal(t, query.Column{
			Text: "B",
			Data: query.NumberColumn{10.0},
		}, response.Columns[2])
	}

	mock.AssertExpectationsForObjects(t, dbh)
}

func TestDeathsByCountry(t *testing.T) {
	dbh := &mockCovidStore.CovidStore{}
	dbh.On("GetAllCountryNames").Return([]string{"A", "B"}, nil)
	dbh.On("GetLatestForCountriesByTime", []string{"A", "B"}, mock.AnythingOfType("time.Time")).Return(map[string]models.CountryEntry{
		"A": {Timestamp: time.Now(), Deaths: 0},
		"B": {Timestamp: time.Now(), Deaths: 1},
	}, nil)
	dbh.On("GetLatestForCountries", []string{"A", "B"}).Return(map[string]models.CountryEntry{
		"A": {Timestamp: time.Now(), Deaths: 0},
		"B": {Timestamp: time.Now(), Deaths: 1},
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
		response, err := h.Endpoints().TableQuery(ctx, args)
		require.NoError(t, err)
		assert.Equal(t, "timestamp", response.Columns[0].Text)
		assert.Len(t, response.Columns[0].Data, 1)
		assert.Equal(t, query.Column{
			Text: "A",
			Data: query.NumberColumn{0.0},
		}, response.Columns[1])
		assert.Equal(t, query.Column{
			Text: "B",
			Data: query.NumberColumn{1.0},
		}, response.Columns[2])
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
	_, err := h.Endpoints().TableQuery(ctx, args)
	assert.Error(t, err)

	dbh.On("GetAllCountryNames").Return([]string{"A", "B"}, nil).Once()
	dbh.On("GetLatestForCountriesByTime", []string{"A", "B"}, mock.AnythingOfType("time.Time")).Return(nil, errors.New("db error")).Once()

	_, err = h.Endpoints().TableQuery(ctx, args)
	assert.Error(t, err)

	mock.AssertExpectationsForObjects(t, dbh)
}
