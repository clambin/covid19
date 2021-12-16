package handler_test

import (
	"context"
	"github.com/clambin/covid19/cache"
	mockCovidStore "github.com/clambin/covid19/covid/store/mocks"
	"github.com/clambin/covid19/handler"
	grafanaJson "github.com/clambin/grafana-json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestEvolution(t *testing.T) {
	dbh := &mockCovidStore.CovidStore{}
	dbh.On("GetAll").Return(dbContents, nil)

	c := &cache.Cache{DB: dbh, Retention: 20 * time.Minute}
	h := handler.Handler{Cache: c}

	args := grafanaJson.TableQueryArgs{
		CommonQueryArgs: grafanaJson.CommonQueryArgs{
			Range: grafanaJson.QueryRequestRange{
				To: time.Now(),
			},
		},
	}

	ctx := context.Background()

	response, err := h.TableQuery(ctx, "evolution", &args)
	require.NoError(t, err)
	require.Len(t, response.Columns, 3)
	assert.Equal(t, "timestamp", response.Columns[0].Text)
	assert.Len(t, response.Columns[0].Data, 2)
	assert.Equal(t, grafanaJson.TableQueryResponseColumn{
		Text: "country",
		Data: grafanaJson.TableQueryResponseStringColumn{"A", "B"},
	}, response.Columns[1])
	assert.Equal(t, grafanaJson.TableQueryResponseColumn{
		Text: "increase",
		Data: grafanaJson.TableQueryResponseNumberColumn{2.0, 7.0},
	}, response.Columns[2])

	mock.AssertExpectationsForObjects(t, dbh)
}
