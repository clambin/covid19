package summarized_test

import (
	"context"
	mockCovidStore "github.com/clambin/covid19/covid/store/mocks"
	"github.com/clambin/covid19/simplejsonserver/summarized"
	"github.com/clambin/simplejson/v3/common"
	"github.com/clambin/simplejson/v3/query"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestIncrementalHandler_Global(t *testing.T) {
	dbh := &mockCovidStore.CovidStore{}
	dbh.On("GetTotalsPerDay").Return(dbTotals, nil)

	h := summarized.IncrementalHandler{Retriever: summarized.Retriever{DB: dbh}}

	args := query.Args{Args: common.Args{Range: common.Range{To: time.Now()}}}

	ctx := context.Background()

	response, err := h.Endpoints().Query(ctx, query.Request{Args: args})
	require.NoError(t, err)
	assert.Equal(t, &query.TableResponse{Columns: []query.Column{
		{Text: "timestamp", Data: query.TimeColumn{
			time.Date(2020, time.November, 1, 0, 0, 0, 0, time.UTC),
			time.Date(2020, time.November, 2, 0, 0, 0, 0, time.UTC),
			time.Date(2020, time.November, 3, 0, 0, 0, 0, time.UTC),
			time.Date(2020, time.November, 4, 0, 0, 0, 0, time.UTC),
		}},
		{Text: "confirmed", Data: query.NumberColumn{1, 2, 0, 7}},
		{Text: "deaths", Data: query.NumberColumn{0, 0, 0, 1}},
	}}, response)

	mock.AssertExpectationsForObjects(t, dbh)
}

func TestIncrementalHandler_Country(t *testing.T) {
	dbh := &mockCovidStore.CovidStore{}
	dbh.On("GetAllForCountryName", "A").Return(filterByName(dbContents, "A"), nil)

	h := summarized.IncrementalHandler{Retriever: summarized.Retriever{DB: dbh}}

	args := query.Args{
		Args: common.Args{
			Range: common.Range{
				To: time.Now(),
			},
			AdHocFilters: []common.AdHocFilter{
				{
					Key:      "Country Name",
					Operator: "=",
					Value:    "A",
				},
			},
		},
	}

	ctx := context.Background()

	response, err := h.Endpoints().Query(ctx, query.Request{Args: args})
	require.NoError(t, err)
	assert.Equal(t, &query.TableResponse{Columns: []query.Column{
		{Text: "timestamp", Data: query.TimeColumn{time.Date(2020, time.November, 1, 0, 0, 0, 0, time.UTC), time.Date(2020, time.November, 2, 0, 0, 0, 0, time.UTC)}},
		{Text: "confirmed", Data: query.NumberColumn{1, 2}},
		{Text: "deaths", Data: query.NumberColumn{0, 0}},
	}}, response)

	mock.AssertExpectationsForObjects(t, dbh)
}

func TestIncrementalHandler_Tags(t *testing.T) {
	dbh := &mockCovidStore.CovidStore{}
	dbh.On("GetAllCountryNames").Return([]string{"A", "B"}, nil)

	h := summarized.IncrementalHandler{Retriever: summarized.Retriever{DB: dbh}}

	ctx := context.Background()

	keys := h.Endpoints().TagKeys(ctx)
	assert.Equal(t, []string{"Country Name"}, keys)

	values, err := h.Endpoints().TagValues(ctx, keys[0])
	require.NoError(t, err)
	assert.Equal(t, []string{"A", "B"}, values)

	mock.AssertExpectationsForObjects(t, dbh)
}
