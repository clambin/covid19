package countries_test

import (
	"context"
	"errors"
	mockCovidStore "github.com/clambin/covid19/db/mocks"
	"github.com/clambin/covid19/models"
	"github.com/clambin/covid19/simplejsonserver/countries"
	"github.com/clambin/simplejson/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestConfirmedByCountry(t *testing.T) {
	dbh := mockCovidStore.NewCovidStore(t)

	timestamp := time.Date(2022, 1, 26, 0, 0, 0, 0, time.UTC)
	dbh.On("GetLatestForCountriesByTime", mock.AnythingOfType("time.Time")).Return(map[string]models.CountryEntry{
		"A": {Timestamp: timestamp, Confirmed: 3},
		"B": {Timestamp: timestamp, Confirmed: 10},
	}, nil)
	dbh.On("GetLatestForCountries").Return(map[string]models.CountryEntry{
		"A": {Timestamp: timestamp, Confirmed: 3},
		"B": {Timestamp: timestamp, Confirmed: 10},
	}, nil)

	h := countries.ByCountryHandler{
		CovidDB: dbh,
		Mode:    countries.CountryConfirmed,
	}

	ctx := context.Background()

	for _, args := range []simplejson.QueryArgs{
		{Args: simplejson.Args{Range: simplejson.Range{To: timestamp}}},
		{},
	} {
		response, err := h.Endpoints().Query(ctx, simplejson.QueryRequest{QueryArgs: args})
		require.NoError(t, err)
		assert.Equal(t, &simplejson.TableResponse{Columns: []simplejson.Column{
			{Text: "timestamp", Data: simplejson.TimeColumn{timestamp}},
			{Text: "A", Data: simplejson.NumberColumn{3}},
			{Text: "B", Data: simplejson.NumberColumn{10}},
		}}, response)
	}
}

func TestDeathsByCountry(t *testing.T) {
	dbh := mockCovidStore.NewCovidStore(t)

	timestamp := time.Date(2022, 1, 26, 0, 0, 0, 0, time.UTC)
	dbh.On("GetLatestForCountriesByTime", mock.AnythingOfType("time.Time")).Return(map[string]models.CountryEntry{
		"A": {Timestamp: timestamp, Deaths: 0},
		"B": {Timestamp: timestamp, Deaths: 1},
	}, nil)
	dbh.On("GetLatestForCountries").Return(map[string]models.CountryEntry{
		"A": {Timestamp: timestamp, Deaths: 0},
		"B": {Timestamp: timestamp, Deaths: 1},
	}, nil)

	h := countries.ByCountryHandler{
		CovidDB: dbh,
		Mode:    countries.CountryDeaths,
	}

	ctx := context.Background()
	for _, args := range []simplejson.QueryArgs{
		{Args: simplejson.Args{Range: simplejson.Range{To: time.Now()}}},
		{},
	} {
		response, err := h.Endpoints().Query(ctx, simplejson.QueryRequest{QueryArgs: args})
		require.NoError(t, err)
		assert.Equal(t, &simplejson.TableResponse{Columns: []simplejson.Column{
			{Text: "timestamp", Data: simplejson.TimeColumn{timestamp}},
			{Text: "A", Data: simplejson.NumberColumn{0}},
			{Text: "B", Data: simplejson.NumberColumn{1}},
		}}, response)
	}
}

func TestConfirmedByCountry_Errors(t *testing.T) {
	dbh := mockCovidStore.NewCovidStore(t)

	h := countries.ByCountryHandler{
		CovidDB: dbh,
		Mode:    countries.CountryConfirmed,
	}

	ctx := context.Background()
	args := simplejson.QueryArgs{Args: simplejson.Args{Range: simplejson.Range{To: time.Now()}}}

	dbh.On("GetLatestForCountriesByTime", mock.AnythingOfType("time.Time")).Return(nil, errors.New("db error")).Once()

	_, err := h.Endpoints().Query(ctx, simplejson.QueryRequest{QueryArgs: args})
	assert.Error(t, err)
}
