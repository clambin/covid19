package mortality_test

import (
	"context"
	"errors"
	mockCovidStore "github.com/clambin/covid19/db/mocks"
	"github.com/clambin/covid19/models"
	"github.com/clambin/covid19/simplejsonserver/mortality"
	"github.com/clambin/simplejson/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestHandler(t *testing.T) {
	dbh := mockCovidStore.NewCovidStore(t)
	dbh.
		On("GetAllCountryNames").
		Return([]string{"AA", "BB", "CC"}, nil)
	dbh.
		On("GetLatestForCountriesByTime", []string{"AA", "BB", "CC"}, mock.AnythingOfType("time.Time")).
		Return(map[string]models.CountryEntry{
			"AA": {
				Timestamp: time.Date(2021, 12, 17, 0, 0, 0, 0, time.UTC),
				Code:      "A",
				Name:      "AA",
				Confirmed: 100,
				Deaths:    10,
			},
			"BB": {
				Timestamp: time.Date(2021, 12, 17, 0, 0, 0, 0, time.UTC),
				Code:      "B",
				Name:      "BB",
				Confirmed: 200,
				Deaths:    10,
			},
		}, nil)

	h := mortality.Handler{CovidDB: dbh}

	args := simplejson.QueryArgs{Args: simplejson.Args{Range: simplejson.Range{To: time.Now()}}}

	ctx := context.Background()

	response, err := h.Endpoints().Query(ctx, simplejson.QueryRequest{QueryArgs: args})
	require.NoError(t, err)
	assert.Equal(t, &simplejson.TableResponse{Columns: []simplejson.Column{
		{Text: "timestamp", Data: simplejson.TimeColumn{time.Date(2021, time.December, 17, 0, 0, 0, 0, time.UTC), time.Date(2021, time.December, 17, 0, 0, 0, 0, time.UTC)}},
		{Text: "country", Data: simplejson.StringColumn{"A", "B"}},
		{Text: "ratio", Data: simplejson.NumberColumn{0.1, 0.05}},
	}}, response)
}

func TestHandler_Errors(t *testing.T) {
	dbh := mockCovidStore.NewCovidStore(t)
	h := mortality.Handler{CovidDB: dbh}

	args := simplejson.QueryArgs{}

	ctx := context.Background()

	dbh.
		On("GetAllCountryNames").
		Return(nil, errors.New("db error")).
		Once()

	_, err := h.Endpoints().Query(ctx, simplejson.QueryRequest{QueryArgs: args})
	require.Error(t, err)

	dbh.
		On("GetAllCountryNames").
		Return([]string{"AA", "BB", "CC"}, nil)
	dbh.
		On("GetLatestForCountriesByTime", []string{"AA", "BB", "CC"}, mock.AnythingOfType("time.Time")).
		Return(nil, errors.New("db error"))

	_, err = h.Endpoints().Query(ctx, simplejson.QueryRequest{QueryArgs: args})
	require.Error(t, err)
}
