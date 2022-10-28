package updates_test

import (
	"context"
	mockCovidStore "github.com/clambin/covid19/db/mocks"
	"github.com/clambin/covid19/simplejsonserver/updates"
	"github.com/clambin/simplejson/v3/common"
	"github.com/clambin/simplejson/v3/query"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestHandler_Updates(t *testing.T) {
	dbh := &mockCovidStore.CovidStore{}
	dbh.
		On("CountEntriesByTime", time.Date(2020, time.November, 2, 0, 0, 0, 0, time.UTC), time.Date(2020, time.November, 3, 0, 0, 0, 0, time.UTC)).
		Return(map[time.Time]int{
			time.Date(2020, time.November, 3, 0, 0, 0, 0, time.UTC): 5,
			time.Date(2020, time.November, 2, 0, 0, 0, 0, time.UTC): 1,
		}, nil)

	h := updates.Handler{CovidDB: dbh}

	args := query.Args{
		Args: common.Args{Range: common.Range{
			From: time.Date(2020, time.November, 2, 0, 0, 0, 0, time.UTC),
			To:   time.Date(2020, time.November, 3, 0, 0, 0, 0, time.UTC),
		}},
	}

	ctx := context.Background()

	response, err := h.Endpoints().Query(ctx, query.Request{Args: args})
	require.NoError(t, err)
	assert.Equal(t, &query.TableResponse{Columns: []query.Column{
		{Text: "timestamp", Data: query.TimeColumn{time.Date(2020, time.November, 2, 0, 0, 0, 0, time.UTC), time.Date(2020, time.November, 3, 0, 0, 0, 0, time.UTC)}},
		{Text: "updates", Data: query.NumberColumn{1, 5}},
	}}, response)

	mock.AssertExpectationsForObjects(t, dbh)
}
