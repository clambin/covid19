package summarized_test

import (
	"context"
	mockCovidStore "github.com/clambin/covid19/db/mocks"
	"github.com/clambin/covid19/simplejsonserver/summarized"
	"github.com/clambin/simplejson/v6"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestIncrementalHandler_Global(t *testing.T) {
	dbh := mockCovidStore.NewCovidStore(t)
	dbh.On("GetTotalsPerDay").Return(dbTotals, nil)

	h := summarized.IncrementalHandler{Retriever: summarized.Retriever{DB: dbh}}

	args := simplejson.QueryArgs{Args: simplejson.Args{Range: simplejson.Range{To: time.Now()}}}

	ctx := context.Background()

	response, err := h.Endpoints().Query(ctx, simplejson.QueryRequest{QueryArgs: args})
	require.NoError(t, err)
	assert.Equal(t, &simplejson.TableResponse{Columns: []simplejson.Column{
		{Text: "timestamp", Data: simplejson.TimeColumn{
			time.Date(2020, time.November, 1, 0, 0, 0, 0, time.UTC),
			time.Date(2020, time.November, 2, 0, 0, 0, 0, time.UTC),
			time.Date(2020, time.November, 3, 0, 0, 0, 0, time.UTC),
			time.Date(2020, time.November, 4, 0, 0, 0, 0, time.UTC),
		}},
		{Text: "confirmed", Data: simplejson.NumberColumn{1, 2, 0, 7}},
		{Text: "deaths", Data: simplejson.NumberColumn{0, 0, 0, 1}},
	}}, response)
}

func TestIncrementalHandler_Country(t *testing.T) {
	dbh := mockCovidStore.NewCovidStore(t)
	dbh.On("GetAllForCountryName", "A").Return(filterByName(dbContents, "A"), nil)

	h := summarized.IncrementalHandler{Retriever: summarized.Retriever{DB: dbh}}

	args := simplejson.QueryArgs{
		Args: simplejson.Args{
			Range: simplejson.Range{
				To: time.Now(),
			},
			AdHocFilters: []simplejson.AdHocFilter{
				{
					Key:      "Country Name",
					Operator: "=",
					Value:    "A",
				},
			},
		},
	}

	ctx := context.Background()

	response, err := h.Endpoints().Query(ctx, simplejson.QueryRequest{QueryArgs: args})
	require.NoError(t, err)
	assert.Equal(t, &simplejson.TableResponse{Columns: []simplejson.Column{
		{Text: "timestamp", Data: simplejson.TimeColumn{time.Date(2020, time.November, 1, 0, 0, 0, 0, time.UTC), time.Date(2020, time.November, 2, 0, 0, 0, 0, time.UTC)}},
		{Text: "confirmed", Data: simplejson.NumberColumn{1, 2}},
		{Text: "deaths", Data: simplejson.NumberColumn{0, 0}},
	}}, response)
}

func TestIncrementalHandler_Tags(t *testing.T) {
	dbh := mockCovidStore.NewCovidStore(t)
	dbh.On("GetAllCountryNames").Return([]string{"A", "B"}, nil)

	h := summarized.IncrementalHandler{Retriever: summarized.Retriever{DB: dbh}}

	ctx := context.Background()

	keys := h.Endpoints().TagKeys(ctx)
	assert.Equal(t, []string{"Country Name"}, keys)

	values, err := h.Endpoints().TagValues(ctx, keys[0])
	require.NoError(t, err)
	assert.Equal(t, []string{"A", "B"}, values)
}
