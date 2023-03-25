package updates_test

import (
	"context"
	"github.com/clambin/covid19/internal/testtools/db/covid"
	"github.com/clambin/covid19/models"
	"github.com/clambin/covid19/simplejsonserver/updates"
	"github.com/clambin/simplejson/v6"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestHandler_Updates(t *testing.T) {
	covidDB := covid.FakeStore{Records: []models.CountryEntry{
		{Timestamp: time.Date(2020, time.November, 2, 0, 0, 0, 0, time.UTC)},
		{Timestamp: time.Date(2020, time.November, 3, 0, 0, 0, 0, time.UTC)},
		{Timestamp: time.Date(2020, time.November, 3, 0, 0, 0, 0, time.UTC)},
		{Timestamp: time.Date(2020, time.November, 3, 0, 0, 0, 0, time.UTC)},
		{Timestamp: time.Date(2020, time.November, 3, 0, 0, 0, 0, time.UTC)},
		{Timestamp: time.Date(2020, time.November, 3, 0, 0, 0, 0, time.UTC)},
	}}
	h := updates.Handler{DB: &covidDB}

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
