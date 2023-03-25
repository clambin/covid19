package evolution_test

import (
	"context"
	"github.com/clambin/covid19/covid"
	covid2 "github.com/clambin/covid19/internal/testtools/db/covid"
	"github.com/clambin/covid19/models"
	"github.com/clambin/covid19/simplejsonserver/evolution"
	"github.com/clambin/simplejson/v6"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestEvolution(t *testing.T) {
	db := covid2.FakeStore{Records: dbContents}
	h := evolution.Handler{CovidDB: &db}
	args := simplejson.QueryArgs{Args: simplejson.Args{Range: simplejson.Range{To: time.Date(2022, time.January, 4, 0, 0, 0, 0, time.UTC)}}}
	ctx := context.Background()

	response, err := h.Endpoints().Query(ctx, simplejson.QueryRequest{QueryArgs: args})
	require.NoError(t, err)
	assert.Equal(t, &simplejson.TableResponse{Columns: []simplejson.Column{
		{Text: "timestamp", Data: simplejson.TimeColumn{time.Date(2022, time.January, 4, 0, 0, 0, 0, time.UTC), time.Date(2022, time.January, 4, 0, 0, 0, 0, time.UTC)}},
		{Text: "country", Data: simplejson.StringColumn{"A", "B"}},
		{Text: "increase", Data: simplejson.NumberColumn{2, 3.5}},
	}}, response)
}

func TestEvolution_NoEndDate(t *testing.T) {
	db := covid2.FakeStore{Records: dbContents}
	h := evolution.Handler{CovidDB: &db}
	args := simplejson.QueryArgs{}
	ctx := context.Background()

	response, err := h.Endpoints().Query(ctx, simplejson.QueryRequest{QueryArgs: args})
	require.NoError(t, err)
	assert.Equal(t, &simplejson.TableResponse{Columns: []simplejson.Column{
		{Text: "timestamp", Data: simplejson.TimeColumn(nil)},
		{Text: "country", Data: simplejson.StringColumn(nil)},
		{Text: "increase", Data: simplejson.NumberColumn(nil)},
	}}, response)
}

func TestEvolution_NoData(t *testing.T) {
	h := evolution.Handler{CovidDB: &covid2.FakeStore{Records: dbContents}}

	args := simplejson.QueryArgs{Args: simplejson.Args{Range: simplejson.Range{To: time.Date(2020, time.October, 31, 0, 0, 0, 0, time.UTC)}}}
	ctx := context.Background()

	response, err := h.Endpoints().Query(ctx, simplejson.QueryRequest{QueryArgs: args})
	require.NoError(t, err)
	assert.Equal(t, &simplejson.TableResponse{Columns: []simplejson.Column{
		{Text: "timestamp", Data: simplejson.TimeColumn(nil)},
		{Text: "country", Data: simplejson.StringColumn(nil)},
		{Text: "increase", Data: simplejson.NumberColumn(nil)},
	}}, response)
}

func BenchmarkHandler_TableQuery_Evolution(b *testing.B) {
	var bigData []models.CountryEntry
	timestamp := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	for i := 0; i < 2*365; i++ {
		for name, code := range covid.CountryCodes {
			bigData = append(bigData, models.CountryEntry{
				Timestamp: timestamp,
				Code:      code,
				Name:      name,
			})
		}
		timestamp.Add(24 * time.Hour)
	}

	h := evolution.Handler{CovidDB: stubbedStore{records: bigData}}
	args := simplejson.QueryArgs{Args: simplejson.Args{Range: simplejson.Range{To: timestamp}}}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := h.Endpoints().Query(ctx, simplejson.QueryRequest{QueryArgs: args})
		if err != nil {
			panic(err)
		}
	}
}

type stubbedStore struct {
	records []models.CountryEntry
}

func (s stubbedStore) GetAllForRange(_, _ time.Time) ([]models.CountryEntry, error) {
	return s.records, nil
}

var dbContents = []models.CountryEntry{
	{
		Timestamp: time.Date(2022, time.January, 1, 0, 0, 0, 0, time.UTC),
		Code:      "A",
		Name:      "A",
		Confirmed: 1,
		Recovered: 0,
		Deaths:    0,
	},
	{
		Timestamp: time.Date(2022, time.January, 2, 0, 0, 0, 0, time.UTC),
		Code:      "B",
		Name:      "B",
		Confirmed: 3,
		Recovered: 0,
		Deaths:    0,
	},
	{
		Timestamp: time.Date(2022, time.January, 2, 0, 0, 0, 0, time.UTC),
		Code:      "A",
		Name:      "A",
		Confirmed: 3,
		Recovered: 1,
		Deaths:    0,
	},
	{
		Timestamp: time.Date(2022, time.January, 4, 0, 0, 0, 0, time.UTC),
		Code:      "B",
		Name:      "B",
		Confirmed: 10,
		Recovered: 5,
		Deaths:    1,
	},
	{
		Timestamp: time.Date(2022, time.January, 5, 0, 0, 0, 0, time.UTC),
		Code:      "B",
		Name:      "B",
		Confirmed: 20,
		Recovered: 15,
		Deaths:    2,
	},
}
