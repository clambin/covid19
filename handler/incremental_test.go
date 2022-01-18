package handler_test

import (
	"context"
	"github.com/clambin/covid19/cache"
	mockCovidStore "github.com/clambin/covid19/covid/store/mocks"
	"github.com/clambin/covid19/handler"
	"github.com/clambin/simplejson"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestIncremental(t *testing.T) {
	dbh := &mockCovidStore.CovidStore{}
	dbh.On("GetAll").Return(dbContents, nil)

	c := &cache.Cache{DB: dbh, Retention: 20 * time.Minute}
	h := handler.Handler{Cache: c}

	args := simplejson.TableQueryArgs{
		Args: simplejson.Args{
			Range: simplejson.Range{
				To: time.Now(),
			},
		},
	}

	ctx := context.Background()

	response, err := h.TableQuery(ctx, "incremental", &args)
	require.NoError(t, err)
	require.Len(t, response.Columns, 3)
	for i := 0; i < 3; i++ {
		require.Len(t, response.Columns[i].Data, 3)
	}
	assert.Equal(t, simplejson.TableQueryResponseNumberColumn{0, 0, 1}, response.Columns[1].Data)
	assert.Equal(t, simplejson.TableQueryResponseNumberColumn{1, 5, 7}, response.Columns[2].Data)

	mock.AssertExpectationsForObjects(t, dbh)
}

func TestIncrementalForCountry(t *testing.T) {
	dbh := &mockCovidStore.CovidStore{}
	dbh.On("GetAllForCountryName", "A").Return(filterByName(dbContents, "A"), nil)

	c := &cache.Cache{DB: dbh, Retention: 20 * time.Minute}
	h := handler.Handler{Cache: c}

	args := simplejson.TableQueryArgs{
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

	response, err := h.TableQuery(ctx, "incremental", &args)
	require.NoError(t, err)
	require.Len(t, response.Columns, 3)
	for i := 0; i < 3; i++ {
		require.Len(t, response.Columns[i].Data, 2)
	}
	assert.Equal(t, simplejson.TableQueryResponseNumberColumn{0, 0}, response.Columns[1].Data)
	assert.Equal(t, simplejson.TableQueryResponseNumberColumn{1, 2}, response.Columns[2].Data)

	mock.AssertExpectationsForObjects(t, dbh)
}
