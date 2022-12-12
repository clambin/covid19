package countries_test

import (
	"context"
	"errors"
	mockCovidStore "github.com/clambin/covid19/db/mocks"
	"github.com/clambin/covid19/models"
	"github.com/clambin/covid19/simplejsonserver/countries"
	"github.com/clambin/simplejson/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestConfirmedByCountryByPopulation(t *testing.T) {
	dbh := mockCovidStore.NewCovidStore(t)
	dbh.
		On("GetAllCountryNames").
		Return([]string{"Belgium", "US"}, nil)
	dbh.
		On("GetLatestForCountriesByTime", []string{"Belgium", "US"}, time.Date(2020, 1, 3, 0, 0, 0, 0, time.UTC)).
		Return(map[string]models.CountryEntry{
			"Belgium": {Timestamp: time.Date(2020, 1, 3, 0, 0, 0, 0, time.UTC), Name: "Belgium", Code: "BE", Confirmed: 200},
			"US":      {Timestamp: time.Date(2020, 1, 3, 0, 0, 0, 0, time.UTC), Name: "US", Code: "US", Confirmed: 200},
		}, nil)

	dbh2 := mockCovidStore.NewPopulationStore(t)
	dbh2.On("List").Return(map[string]int64{
		"BE": 10,
		"US": 20,
	}, nil)

	h := countries.ByCountryByPopulationHandler{
		CovidDB: dbh,
		PopDB:   dbh2,
		Mode:    countries.CountryConfirmed,
	}

	args := simplejson.QueryArgs{Args: simplejson.Args{Range: simplejson.Range{To: time.Date(2020, 1, 3, 0, 0, 0, 0, time.UTC)}}}
	ctx := context.Background()

	response, err := h.Endpoints().Query(ctx, simplejson.QueryRequest{QueryArgs: args})
	require.NoError(t, err)
	assert.Equal(t, &simplejson.TableResponse{Columns: []simplejson.Column{
		{Text: "timestamp", Data: simplejson.TimeColumn{time.Date(2020, time.January, 3, 0, 0, 0, 0, time.UTC), time.Date(2020, time.January, 3, 0, 0, 0, 0, time.UTC)}},
		{Text: "country", Data: simplejson.StringColumn{"BE", "US"}},
		{Text: "confirmed", Data: simplejson.NumberColumn{20, 10}},
	}}, response)
}

func TestDeathsByCountryByPopulation(t *testing.T) {
	dbh := mockCovidStore.NewCovidStore(t)
	dbh.
		On("GetAllCountryNames").
		Return([]string{"Belgium", "US"}, nil)
	dbh.
		On("GetLatestForCountriesByTime", []string{"Belgium", "US"}, time.Date(2020, 1, 3, 0, 0, 0, 0, time.UTC)).
		Return(map[string]models.CountryEntry{
			"Belgium": {Timestamp: time.Date(2020, 1, 3, 0, 0, 0, 0, time.UTC), Name: "Belgium", Code: "BE", Deaths: 200},
			"US":      {Timestamp: time.Date(2020, 1, 3, 0, 0, 0, 0, time.UTC), Name: "US", Code: "US", Deaths: 200},
		}, nil)

	dbh2 := mockCovidStore.NewPopulationStore(t)
	dbh2.On("List").Return(map[string]int64{
		"BE": 10,
		"US": 20,
	}, nil)

	h := countries.ByCountryByPopulationHandler{
		CovidDB: dbh,
		PopDB:   dbh2,
		Mode:    countries.CountryDeaths,
	}

	args := simplejson.QueryArgs{Args: simplejson.Args{Range: simplejson.Range{To: time.Date(2020, 1, 3, 0, 0, 0, 0, time.UTC)}}}
	ctx := context.Background()

	response, err := h.Endpoints().Query(ctx, simplejson.QueryRequest{QueryArgs: args})
	require.NoError(t, err)
	assert.Equal(t, &simplejson.TableResponse{Columns: []simplejson.Column{
		{Text: "timestamp", Data: simplejson.TimeColumn{time.Date(2020, time.January, 3, 0, 0, 0, 0, time.UTC), time.Date(2020, time.January, 3, 0, 0, 0, 0, time.UTC)}},
		{Text: "country", Data: simplejson.StringColumn{"BE", "US"}},
		{Text: "deaths", Data: simplejson.NumberColumn{20, 10}},
	}}, response)
}

func TestConfirmedByCountryByPopulation_Errors(t *testing.T) {
	dbh := mockCovidStore.NewCovidStore(t)
	dbh.
		On("GetAllCountryNames").
		Return([]string{"Belgium", "US"}, nil)
	dbh.
		On("GetLatestForCountriesByTime", []string{"Belgium", "US"}, time.Date(2020, 1, 3, 0, 0, 0, 0, time.UTC)).
		Return(map[string]models.CountryEntry{
			"Belgium": {Timestamp: time.Date(2020, 1, 3, 0, 0, 0, 0, time.UTC), Name: "Belgium", Code: "BE", Confirmed: 200},
			"US":      {Timestamp: time.Date(2020, 1, 3, 0, 0, 0, 0, time.UTC), Name: "US", Code: "US", Confirmed: 200},
		}, nil)

	dbh2 := mockCovidStore.NewPopulationStore(t)

	h := countries.ByCountryByPopulationHandler{
		CovidDB: dbh,
		PopDB:   dbh2,
		Mode:    countries.CountryConfirmed,
	}

	args := simplejson.QueryArgs{Args: simplejson.Args{Range: simplejson.Range{To: time.Date(2020, 1, 3, 0, 0, 0, 0, time.UTC)}}}
	ctx := context.Background()

	dbh2.On("List").Return(nil, errors.New("db error"))
	_, err := h.Endpoints().Query(ctx, simplejson.QueryRequest{QueryArgs: args})
	assert.Error(t, err)
}
