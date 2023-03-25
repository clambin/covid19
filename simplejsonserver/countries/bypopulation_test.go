package countries_test

import (
	"context"
	covid2 "github.com/clambin/covid19/internal/testtools/db/covid"
	"github.com/clambin/covid19/internal/testtools/db/population"
	"github.com/clambin/covid19/models"
	"github.com/clambin/covid19/simplejsonserver/countries"
	"github.com/clambin/simplejson/v6"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestConfirmedByCountryByPopulation(t *testing.T) {
	db := covid2.FakeStore{Records: []models.CountryEntry{
		{Timestamp: time.Date(2020, 1, 3, 0, 0, 0, 0, time.UTC), Name: "Belgium", Code: "BE", Confirmed: 200},
		{Timestamp: time.Date(2020, 1, 3, 0, 0, 0, 0, time.UTC), Name: "US", Code: "US", Confirmed: 200},
	}}

	db2 := population.FakeStore{}
	_ = db2.Add("BE", 10)
	_ = db2.Add("US", 20)

	h := countries.ByCountryByPopulationHandler{
		CovidDB: &db,
		PopDB:   &db2,
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
	db := covid2.FakeStore{Records: []models.CountryEntry{
		{Timestamp: time.Date(2020, 1, 3, 0, 0, 0, 0, time.UTC), Name: "Belgium", Code: "BE", Deaths: 200},
		{Timestamp: time.Date(2020, 1, 3, 0, 0, 0, 0, time.UTC), Name: "US", Code: "US", Deaths: 200},
	}}

	db2 := population.FakeStore{}
	_ = db2.Add("BE", 10)
	_ = db2.Add("US", 20)

	h := countries.ByCountryByPopulationHandler{
		CovidDB: &db,
		PopDB:   &db2,
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
	db := covid2.FakeStore{Records: []models.CountryEntry{
		{Timestamp: time.Date(2020, 1, 3, 0, 0, 0, 0, time.UTC), Name: "Belgium", Code: "BE", Confirmed: 200},
		{Timestamp: time.Date(2020, 1, 3, 0, 0, 0, 0, time.UTC), Name: "US", Code: "US", Confirmed: 200},
	}}

	db2 := population.FakeStore{Fail: true}

	h := countries.ByCountryByPopulationHandler{
		CovidDB: &db,
		PopDB:   &db2,
		Mode:    countries.CountryConfirmed,
	}

	args := simplejson.QueryArgs{Args: simplejson.Args{Range: simplejson.Range{To: time.Date(2020, 1, 3, 0, 0, 0, 0, time.UTC)}}}
	ctx := context.Background()

	_, err := h.Endpoints().Query(ctx, simplejson.QueryRequest{QueryArgs: args})
	assert.Error(t, err)
}
