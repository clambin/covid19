package mortality_test

import (
	"context"
	"errors"
	mockCovidStore "github.com/clambin/covid19/db/mocks"
	"github.com/clambin/covid19/models"
	"github.com/clambin/covid19/simplejsonserver/mortality"
	"github.com/clambin/simplejson/v3/common"
	"github.com/clambin/simplejson/v3/query"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestHandler(t *testing.T) {
	dbh := &mockCovidStore.CovidStore{}
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

	args := query.Args{Args: common.Args{Range: common.Range{To: time.Now()}}}

	ctx := context.Background()

	response, err := h.Endpoints().Query(ctx, query.Request{Args: args})
	require.NoError(t, err)
	assert.Equal(t, &query.TableResponse{Columns: []query.Column{
		{Text: "timestamp", Data: query.TimeColumn{time.Date(2021, time.December, 17, 0, 0, 0, 0, time.UTC), time.Date(2021, time.December, 17, 0, 0, 0, 0, time.UTC)}},
		{Text: "country", Data: query.StringColumn{"A", "B"}},
		{Text: "ratio", Data: query.NumberColumn{0.1, 0.05}},
	}}, response)

	mock.AssertExpectationsForObjects(t, dbh)
}

func TestHandler_Errors(t *testing.T) {
	dbh := &mockCovidStore.CovidStore{}
	h := mortality.Handler{CovidDB: dbh}

	args := query.Args{}

	ctx := context.Background()

	dbh.
		On("GetAllCountryNames").
		Return(nil, errors.New("db error")).
		Once()

	_, err := h.Endpoints().Query(ctx, query.Request{Args: args})
	require.Error(t, err)

	dbh.
		On("GetAllCountryNames").
		Return([]string{"AA", "BB", "CC"}, nil)
	dbh.
		On("GetLatestForCountriesByTime", []string{"AA", "BB", "CC"}, mock.AnythingOfType("time.Time")).
		Return(nil, errors.New("db error"))

	_, err = h.Endpoints().Query(ctx, query.Request{Args: args})
	require.Error(t, err)

	mock.AssertExpectationsForObjects(t, dbh)
}
