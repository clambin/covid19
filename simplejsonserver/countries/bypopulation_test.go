package countries_test

import (
	"context"
	"errors"
	mockCovidStore "github.com/clambin/covid19/covid/store/mocks"
	"github.com/clambin/covid19/models"
	mockPopulationStore "github.com/clambin/covid19/population/store/mocks"
	"github.com/clambin/covid19/simplejsonserver/countries"
	"github.com/clambin/simplejson/v2/common"
	"github.com/clambin/simplejson/v2/query"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestConfirmedByCountryByPopulation(t *testing.T) {
	dbh := &mockCovidStore.CovidStore{}
	dbh.
		On("GetAllCountryNames").
		Return([]string{"Belgium", "US"}, nil)
	dbh.
		On("GetLatestForCountriesByTime", []string{"Belgium", "US"}, time.Date(2020, 1, 3, 0, 0, 0, 0, time.UTC)).
		Return(map[string]models.CountryEntry{
			"Belgium": {Timestamp: time.Date(2020, 1, 3, 0, 0, 0, 0, time.UTC), Name: "Belgium", Code: "BE", Confirmed: 200},
			"US":      {Timestamp: time.Date(2020, 1, 3, 0, 0, 0, 0, time.UTC), Name: "US", Code: "US", Confirmed: 200},
		}, nil)

	dbh2 := &mockPopulationStore.PopulationStore{}
	dbh2.On("List").Return(map[string]int64{
		"BE": 10,
		"US": 20,
	}, nil)

	h := countries.ByCountryByPopulationHandler{
		CovidDB: dbh,
		PopDB:   dbh2,
		Mode:    countries.CountryConfirmed,
	}

	args := query.Args{Args: common.Args{Range: common.Range{To: time.Date(2020, 1, 3, 0, 0, 0, 0, time.UTC)}}}
	ctx := context.Background()

	response, err := h.Endpoints().TableQuery(ctx, args)
	require.NoError(t, err)
	require.Len(t, response.Columns, 3)
	assert.Equal(t, "timestamp", response.Columns[0].Text)
	assert.Len(t, response.Columns[0].Data, 2)
	assert.Equal(t, query.Column{
		Text: "country",
		Data: query.StringColumn{"BE", "US"},
	}, response.Columns[1])
	assert.Equal(t, query.Column{
		Text: "confirmed",
		Data: query.NumberColumn{20.0, 10.0},
	}, response.Columns[2])

	mock.AssertExpectationsForObjects(t, dbh, dbh2)
}

func TestDeathsByCountryByPopulation(t *testing.T) {
	dbh := &mockCovidStore.CovidStore{}
	dbh.
		On("GetAllCountryNames").
		Return([]string{"Belgium", "US"}, nil)
	dbh.
		On("GetLatestForCountriesByTime", []string{"Belgium", "US"}, time.Date(2020, 1, 3, 0, 0, 0, 0, time.UTC)).
		Return(map[string]models.CountryEntry{
			"Belgium": {Timestamp: time.Date(2020, 1, 3, 0, 0, 0, 0, time.UTC), Name: "Belgium", Code: "BE", Deaths: 200},
			"US":      {Timestamp: time.Date(2020, 1, 3, 0, 0, 0, 0, time.UTC), Name: "US", Code: "US", Deaths: 200},
		}, nil)

	dbh2 := &mockPopulationStore.PopulationStore{}
	dbh2.On("List").Return(map[string]int64{
		"BE": 10,
		"US": 20,
	}, nil)

	h := countries.ByCountryByPopulationHandler{
		CovidDB: dbh,
		PopDB:   dbh2,
		Mode:    countries.CountryDeaths,
	}

	args := query.Args{Args: common.Args{Range: common.Range{To: time.Date(2020, 1, 3, 0, 0, 0, 0, time.UTC)}}}
	ctx := context.Background()

	response, err := h.Endpoints().TableQuery(ctx, args)
	require.NoError(t, err)
	require.Len(t, response.Columns, 3)
	assert.Equal(t, "timestamp", response.Columns[0].Text)
	assert.Len(t, response.Columns[0].Data, 2)
	assert.Equal(t, query.Column{
		Text: "country",
		Data: query.StringColumn{"BE", "US"},
	}, response.Columns[1])
	assert.Equal(t, query.Column{
		Text: "deaths",
		Data: query.NumberColumn{20.0, 10.0},
	}, response.Columns[2])

	mock.AssertExpectationsForObjects(t, dbh, dbh2)
}

func TestConfirmedByCountryByPopulation_Errors(t *testing.T) {
	dbh := &mockCovidStore.CovidStore{}
	dbh.
		On("GetAllCountryNames").
		Return([]string{"Belgium", "US"}, nil)
	dbh.
		On("GetLatestForCountriesByTime", []string{"Belgium", "US"}, time.Date(2020, 1, 3, 0, 0, 0, 0, time.UTC)).
		Return(map[string]models.CountryEntry{
			"Belgium": {Timestamp: time.Date(2020, 1, 3, 0, 0, 0, 0, time.UTC), Name: "Belgium", Code: "BE", Confirmed: 200},
			"US":      {Timestamp: time.Date(2020, 1, 3, 0, 0, 0, 0, time.UTC), Name: "US", Code: "US", Confirmed: 200},
		}, nil)

	dbh2 := &mockPopulationStore.PopulationStore{}

	h := countries.ByCountryByPopulationHandler{
		CovidDB: dbh,
		PopDB:   dbh2,
		Mode:    countries.CountryConfirmed,
	}

	args := query.Args{Args: common.Args{Range: common.Range{To: time.Date(2020, 1, 3, 0, 0, 0, 0, time.UTC)}}}
	ctx := context.Background()

	dbh2.On("List").Return(nil, errors.New("db error"))
	_, err := h.Endpoints().TableQuery(ctx, args)
	assert.Error(t, err)

	mock.AssertExpectationsForObjects(t, dbh, dbh2)
}
