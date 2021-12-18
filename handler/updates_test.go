package handler_test

import (
	"context"
	"github.com/clambin/covid19/cache"
	mockCovidStore "github.com/clambin/covid19/covid/store/mocks"
	"github.com/clambin/covid19/handler"
	grafanaJson "github.com/clambin/grafana-json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestHandler_Updates(t *testing.T) {
	dbh := &mockCovidStore.CovidStore{}
	dbh.
		On("GetAll").
		Return(dbContents, nil)

	c := &cache.Cache{DB: dbh, Retention: 20 * time.Minute}
	h := handler.Handler{Cache: c}

	args := grafanaJson.TableQueryArgs{}

	ctx := context.Background()

	response, err := h.TableQuery(ctx, "updates", &args)
	require.NoError(t, err)
	require.Len(t, response.Columns, 2)
	assert.Equal(t, "timestamp", response.Columns[0].Text)
	assert.Equal(t, grafanaJson.TableQueryResponseTimeColumn{
		time.Date(2020, time.November, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2020, time.November, 2, 0, 0, 0, 0, time.UTC),
		time.Date(2020, time.November, 4, 0, 0, 0, 0, time.UTC),
	}, response.Columns[0].Data)
	assert.Equal(t, "updates", response.Columns[1].Text)
	assert.Equal(t, grafanaJson.TableQueryResponseNumberColumn{1, 2, 1}, response.Columns[1].Data)

	args = grafanaJson.TableQueryArgs{
		CommonQueryArgs: grafanaJson.CommonQueryArgs{
			Range: grafanaJson.QueryRequestRange{
				To: time.Date(2020, time.November, 2, 0, 0, 0, 0, time.UTC),
			},
		},
	}
	response, err = h.TableQuery(ctx, "updates", &args)
	require.NoError(t, err)
	require.Len(t, response.Columns, 2)
	assert.Equal(t, "timestamp", response.Columns[0].Text)
	assert.Equal(t, grafanaJson.TableQueryResponseTimeColumn{
		time.Date(2020, time.November, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2020, time.November, 2, 0, 0, 0, 0, time.UTC),
	}, response.Columns[0].Data)
	assert.Equal(t, "updates", response.Columns[1].Text)
	assert.Equal(t, grafanaJson.TableQueryResponseNumberColumn{1, 2}, response.Columns[1].Data)
}
