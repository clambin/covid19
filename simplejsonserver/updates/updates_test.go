package updates_test

import (
	"context"
	mockCovidStore "github.com/clambin/covid19/db/mocks"
	"github.com/clambin/covid19/simplejsonserver/updates"
	"github.com/clambin/simplejson/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestHandler_Updates(t *testing.T) {
	dbh := mockCovidStore.NewCovidStore(t)
	dbh.
		On("CountEntriesByTime", time.Date(2020, time.November, 2, 0, 0, 0, 0, time.UTC), time.Date(2020, time.November, 3, 0, 0, 0, 0, time.UTC)).
		Return([]struct {
			Timestamp time.Time
			Count     int
		}{
			{Timestamp: time.Date(2020, time.November, 2, 0, 0, 0, 0, time.UTC), Count: 1},
			{Timestamp: time.Date(2020, time.November, 3, 0, 0, 0, 0, time.UTC), Count: 5},
		}, nil)

	h := updates.Handler{CovidDB: dbh}

	args := simplejson.QueryArgs{
		Args: simplejson.Args{Range: simplejson.Range{
			From: time.Date(2020, time.November, 2, 0, 0, 0, 0, time.UTC),
			To:   time.Date(2020, time.November, 3, 0, 0, 0, 0, time.UTC),
		}},
	}

	ctx := context.Background()

	response, err := h.Endpoints().Query(ctx, simplejson.QueryRequest{QueryArgs: args})
	require.NoError(t, err)
	assert.Equal(t, &simplejson.TableResponse{Columns: []simplejson.Column{
		{Text: "timestamp", Data: simplejson.TimeColumn{time.Date(2020, time.November, 2, 0, 0, 0, 0, time.UTC), time.Date(2020, time.November, 3, 0, 0, 0, 0, time.UTC)}},
		{Text: "updates", Data: simplejson.NumberColumn{1, 5}},
	}}, response)
}
