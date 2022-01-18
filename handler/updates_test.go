package handler_test

import (
	"context"
	"github.com/clambin/covid19/cache"
	mockCovidStore "github.com/clambin/covid19/covid/store/mocks"
	"github.com/clambin/covid19/handler"
	"github.com/clambin/covid19/models"
	"github.com/clambin/simplejson"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sort"
	"testing"
	"time"
)

func TestHandler_Updates(t *testing.T) {
	dbContents2 := append(dbContents, models.CountryEntry{
		Timestamp: time.Date(2020, time.November, 3, 0, 0, 0, 0, time.UTC),
		Code:      "A",
		Name:      "A",
		Confirmed: 3,
		Recovered: 1,
		Deaths:    0,
	})
	sort.Slice(dbContents2, func(i, j int) bool { return dbContents2[i].Timestamp.Before(dbContents2[j].Timestamp) })

	dbh := &mockCovidStore.CovidStore{}
	dbh.
		On("GetAll").
		Return(dbContents2, nil)

	c := &cache.Cache{DB: dbh, Retention: 20 * time.Minute}
	h := handler.Handler{Cache: c}

	args := simplejson.TableQueryArgs{}

	ctx := context.Background()

	response, err := h.TableQuery(ctx, "updates", &args)
	require.NoError(t, err)
	require.Len(t, response.Columns, 2)
	assert.Equal(t, "timestamp", response.Columns[0].Text)
	assert.Equal(t, simplejson.TableQueryResponseTimeColumn{
		time.Date(2020, time.November, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2020, time.November, 2, 0, 0, 0, 0, time.UTC),
		time.Date(2020, time.November, 4, 0, 0, 0, 0, time.UTC),
	}, response.Columns[0].Data)
	assert.Equal(t, "updates", response.Columns[1].Text)
	assert.Equal(t, simplejson.TableQueryResponseNumberColumn{1, 2, 1}, response.Columns[1].Data)

	args = simplejson.TableQueryArgs{
		Args: simplejson.Args{
			Range: simplejson.Range{
				To: time.Date(2020, time.November, 2, 0, 0, 0, 0, time.UTC),
			},
		},
	}
	response, err = h.TableQuery(ctx, "updates", &args)
	require.NoError(t, err)
	require.Len(t, response.Columns, 2)
	assert.Equal(t, "timestamp", response.Columns[0].Text)
	assert.Equal(t, simplejson.TableQueryResponseTimeColumn{
		time.Date(2020, time.November, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2020, time.November, 2, 0, 0, 0, 0, time.UTC),
	}, response.Columns[0].Data)
	assert.Equal(t, "updates", response.Columns[1].Text)
	assert.Equal(t, simplejson.TableQueryResponseNumberColumn{1, 2}, response.Columns[1].Data)
}
