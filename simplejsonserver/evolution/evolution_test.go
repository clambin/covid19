package evolution_test

import (
	"context"
	"github.com/clambin/covid19/covid/fetcher"
	mockCovidStore "github.com/clambin/covid19/db/mocks"
	"github.com/clambin/covid19/models"
	"github.com/clambin/covid19/simplejsonserver/evolution"
	"github.com/clambin/simplejson/v3/common"
	"github.com/clambin/simplejson/v3/query"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestEvolution(t *testing.T) {
	dbh := &mockCovidStore.CovidStore{}
	dbh.
		On(
			"GetAllForRange",
			dbContents[len(dbContents)-1].Timestamp.Add(-7*24*time.Hour),
			dbContents[len(dbContents)-1].Timestamp,
		).
		Return(dbContents, nil)

	h := evolution.Handler{CovidDB: dbh}

	args := query.Args{Args: common.Args{Range: common.Range{To: dbContents[len(dbContents)-1].Timestamp}}}

	ctx := context.Background()

	response, err := h.Endpoints().Query(ctx, query.Request{Args: args})
	require.NoError(t, err)
	assert.Equal(t, &query.TableResponse{Columns: []query.Column{
		{Text: "timestamp", Data: query.TimeColumn{time.Date(2022, time.January, 4, 0, 0, 0, 0, time.UTC), time.Date(2022, time.January, 4, 0, 0, 0, 0, time.UTC)}},
		{Text: "country", Data: query.StringColumn{"A", "B"}},
		{Text: "increase", Data: query.NumberColumn{2, 3.5}},
	}}, response)

	mock.AssertExpectationsForObjects(t, dbh)
}

func TestEvolution_NoEndDate(t *testing.T) {
	dbh := &mockCovidStore.CovidStore{}
	h := evolution.Handler{CovidDB: dbh}

	args := query.Args{}
	ctx := context.Background()

	dbContents2 := []models.CountryEntry{
		{
			Timestamp: time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC),
			Code:      "A",
			Name:      "A",
			Confirmed: 1,
			Recovered: 0,
			Deaths:    0,
		},
	}
	dbContents2 = append(dbContents2, dbContents...)
	dbh.
		On("GetAll").
		Return(dbContents2, nil)

	response, err := h.Endpoints().Query(ctx, query.Request{Args: args})
	require.NoError(t, err)
	assert.Equal(t, &query.TableResponse{Columns: []query.Column{
		{Text: "timestamp", Data: query.TimeColumn{time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC), time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC)}},
		{Text: "country", Data: query.StringColumn{"A", "B"}},
		{Text: "increase", Data: query.NumberColumn{2, 3.5}},
	}}, response)

	mock.AssertExpectationsForObjects(t, dbh)
}

func TestEvolution_NoData(t *testing.T) {
	dbh := &mockCovidStore.CovidStore{}
	dbh.
		On(
			"GetAllForRange",
			time.Date(2020, time.October, 24, 0, 0, 0, 0, time.UTC),
			time.Date(2020, time.October, 31, 0, 0, 0, 0, time.UTC),
		).
		Return([]models.CountryEntry{}, nil)

	h := evolution.Handler{CovidDB: dbh}

	args := query.Args{Args: common.Args{Range: common.Range{To: time.Date(2020, time.October, 31, 0, 0, 0, 0, time.UTC)}}}
	ctx := context.Background()

	response, err := h.Endpoints().Query(ctx, query.Request{Args: args})
	require.NoError(t, err)
	assert.Equal(t, &query.TableResponse{Columns: []query.Column{
		{Text: "timestamp", Data: query.TimeColumn(nil)},
		{Text: "country", Data: query.StringColumn(nil)},
		{Text: "increase", Data: query.NumberColumn(nil)},
	}}, response)

	mock.AssertExpectationsForObjects(t, dbh)
}

func BenchmarkHandler_TableQuery_Evolution(b *testing.B) {
	var bigData []models.CountryEntry
	timestamp := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	for i := 0; i < 2*365; i++ {
		for name, code := range fetcher.CountryCodes {
			bigData = append(bigData, models.CountryEntry{
				Timestamp: timestamp,
				Code:      code,
				Name:      name,
			})
		}
		timestamp.Add(24 * time.Hour)
	}

	dbh := &mockCovidStore.CovidStore{}
	dbh.On("GetAllForRange", timestamp.Add(-7*24*time.Hour), timestamp).Return(bigData, nil)

	h := evolution.Handler{CovidDB: dbh}

	args := query.Args{Args: common.Args{Range: common.Range{To: timestamp}}}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := h.Endpoints().Query(ctx, query.Request{Args: args})
		if err != nil {
			panic(err)
		}
	}

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
}
