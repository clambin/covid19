package mortality_test

import (
	"context"
	"github.com/clambin/covid19/internal/testtools/db/covid"
	"github.com/clambin/covid19/models"
	"github.com/clambin/covid19/simplejsonserver/mortality"
	"github.com/clambin/simplejson/v6"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestHandler(t *testing.T) {
	db := covid.FakeStore{}
	_ = db.Add(
		[]models.CountryEntry{
			{
				Timestamp: time.Date(2021, 12, 17, 0, 0, 0, 0, time.UTC),
				Code:      "A",
				Name:      "AA",
				Confirmed: 100,
				Deaths:    10,
			},
			{
				Timestamp: time.Date(2021, 12, 17, 0, 0, 0, 0, time.UTC),
				Code:      "B",
				Name:      "BB",
				Confirmed: 200,
				Deaths:    10,
			},
		})
	h := mortality.Handler{CovidDB: &db}

	args := simplejson.QueryArgs{Args: simplejson.Args{Range: simplejson.Range{To: time.Now()}}}
	ctx := context.Background()

	response, err := h.Endpoints().Query(ctx, simplejson.QueryRequest{QueryArgs: args})
	require.NoError(t, err)
	assert.Equal(t, &simplejson.TableResponse{Columns: []simplejson.Column{
		{Text: "timestamp", Data: simplejson.TimeColumn{time.Date(2021, time.December, 17, 0, 0, 0, 0, time.UTC), time.Date(2021, time.December, 17, 0, 0, 0, 0, time.UTC)}},
		{Text: "country", Data: simplejson.StringColumn{"A", "B"}},
		{Text: "ratio", Data: simplejson.NumberColumn{0.1, 0.05}},
	}}, response)
}

func TestHandler_Errors(t *testing.T) {
	db := covid.FakeStore{Fail: true}
	h := mortality.Handler{CovidDB: &db}
	args := simplejson.QueryArgs{}
	ctx := context.Background()

	_, err := h.Endpoints().Query(ctx, simplejson.QueryRequest{QueryArgs: args})
	require.Error(t, err)
}
