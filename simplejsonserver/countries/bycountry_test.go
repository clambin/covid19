package countries_test

import (
	"context"
	covid2 "github.com/clambin/covid19/internal/testtools/db/covid"
	"github.com/clambin/covid19/models"
	"github.com/clambin/covid19/simplejsonserver/countries"
	"github.com/clambin/simplejson/v6"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestConfirmedByCountry(t *testing.T) {
	timestamp := time.Date(2022, 1, 26, 0, 0, 0, 0, time.UTC)
	h := countries.ByCountryHandler{
		DB: &covid2.FakeStore{Records: []models.CountryEntry{
			{Timestamp: timestamp, Name: "A", Code: "AA", Confirmed: 3, Recovered: 0, Deaths: 0},
			{Timestamp: timestamp, Name: "B", Code: "BB", Confirmed: 10, Recovered: 1, Deaths: 1},
		}},
		Mode: countries.CountryConfirmed,
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
	timestamp := time.Date(2022, 1, 26, 0, 0, 0, 0, time.UTC)
	h := countries.ByCountryHandler{
		DB: &covid2.FakeStore{Records: []models.CountryEntry{
			{Timestamp: timestamp, Name: "A", Code: "AA", Confirmed: 3, Recovered: 0, Deaths: 0},
			{Timestamp: timestamp, Name: "B", Code: "BB", Confirmed: 10, Recovered: 1, Deaths: 1},
		}},
		Mode: countries.CountryDeaths,
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
	h := countries.ByCountryHandler{
		DB:   &covid2.FakeStore{Fail: true},
		Mode: countries.CountryConfirmed,
	}

	ctx := context.Background()
	args := simplejson.QueryArgs{Args: simplejson.Args{Range: simplejson.Range{To: time.Now()}}}

	_, err := h.Endpoints().Query(ctx, simplejson.QueryRequest{QueryArgs: args})
	assert.Error(t, err)
}
