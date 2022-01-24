package updates_test

import (
	"context"
	mockCovidStore "github.com/clambin/covid19/covid/store/mocks"
	"github.com/clambin/covid19/simplejsonserver/updates"
	"github.com/clambin/simplejson/v2/common"
	"github.com/clambin/simplejson/v2/query"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestHandler_Updates(t *testing.T) {
	dbh := &mockCovidStore.CovidStore{}
	dbh.
		On("CountEntriesByTime", time.Date(2020, time.November, 2, 0, 0, 0, 0, time.UTC), time.Date(2020, time.November, 3, 0, 0, 0, 0, time.UTC)).
		Return(map[time.Time]int{
			time.Date(2020, time.November, 2, 0, 0, 0, 0, time.UTC): 1,
			time.Date(2020, time.November, 3, 0, 0, 0, 0, time.UTC): 5,
		}, nil)

	h := updates.Handler{CovidDB: dbh}

	args := query.Args{
		Args: common.Args{Range: common.Range{
			From: time.Date(2020, time.November, 2, 0, 0, 0, 0, time.UTC),
			To:   time.Date(2020, time.November, 3, 0, 0, 0, 0, time.UTC),
		}},
	}

	ctx := context.Background()

	response, err := h.Endpoints().TableQuery(ctx, args)
	require.NoError(t, err)
	require.Len(t, response.Columns, 2)
	assert.Equal(t, "timestamp", response.Columns[0].Text)
	assert.Equal(t, query.TimeColumn{
		time.Date(2020, time.November, 2, 0, 0, 0, 0, time.UTC),
		time.Date(2020, time.November, 3, 0, 0, 0, 0, time.UTC),
	}, response.Columns[0].Data)
	assert.Equal(t, "updates", response.Columns[1].Text)
	assert.Equal(t, query.NumberColumn{1, 5}, response.Columns[1].Data)
}
