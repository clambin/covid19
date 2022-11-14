package summarized_test

import (
	"context"
	"errors"
	mockCovidStore "github.com/clambin/covid19/db/mocks"
	"github.com/clambin/covid19/models"
	"github.com/clambin/covid19/simplejsonserver/summarized"
	"github.com/clambin/simplejson/v3/common"
	"github.com/clambin/simplejson/v3/query"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"strconv"
	"testing"
	"time"
)

func TestCumulativeHandler_Global(t *testing.T) {
	dbh := mockCovidStore.NewCovidStore(t)
	dbh.On("GetTotalsPerDay").Return(dbTotals, nil)

	h := summarized.CumulativeHandler{Retriever: summarized.Retriever{DB: dbh}}

	args := query.Args{Args: common.Args{Range: common.Range{To: time.Now()}}}

	ctx := context.Background()

	response, err := h.Endpoints().Query(ctx, query.Request{Args: args})
	require.NoError(t, err)
	assert.Equal(t, &query.TableResponse{Columns: []query.Column{
		{Text: "timestamp", Data: query.TimeColumn{
			time.Date(2020, time.November, 1, 0, 0, 0, 0, time.UTC),
			time.Date(2020, time.November, 2, 0, 0, 0, 0, time.UTC),
			time.Date(2020, time.November, 3, 0, 0, 0, 0, time.UTC),
			time.Date(2020, time.November, 4, 0, 0, 0, 0, time.UTC),
		}},
		{Text: "confirmed", Data: query.NumberColumn{1, 3, 3, 10}},
		{Text: "deaths", Data: query.NumberColumn{0, 0, 0, 1}},
	}}, response)
}

func BenchmarkCumulativeHandler_Global(b *testing.B) {
	from, to, bigContents := buildBigDatabase()
	dbh := &mockCovidStore.CovidStore{}
	dbh.On("GetTotalsPerDay").Return(bigContents, nil)

	h := summarized.CumulativeHandler{Retriever: summarized.Retriever{DB: dbh}}

	args := query.Args{
		Args: common.Args{
			Range: common.Range{
				From: from,
				To:   to,
			},
		},
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := h.Endpoints().Query(ctx, query.Request{Args: args})
		if err != nil {
			b.Fatal(err)
		}
	}
	//mock.AssertExpectationsForObjects(b, dbh)
}

func buildBigDatabase() (from, to time.Time, content []models.CountryEntry) {
	from = time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC)
	timestamp := from
	for i := 0; i < 2*365; i++ {
		to = timestamp
		for c := 0; c < 193; c++ {
			country := strconv.Itoa(c)
			content = append(content, models.CountryEntry{
				Timestamp: timestamp,
				Code:      country,
				Name:      country,
			})
		}
		timestamp = timestamp.Add(24 * time.Hour)
	}
	return
}

func TestCumulativeHandler_Country(t *testing.T) {
	dbh := &mockCovidStore.CovidStore{}
	dbh.On("GetAllForCountryName", "A").Return(filterByName(dbContents, "A"), nil)

	h := summarized.CumulativeHandler{Retriever: summarized.Retriever{DB: dbh}}

	args := query.Args{
		Args: common.Args{
			Range: common.Range{
				To: time.Now(),
			},
			AdHocFilters: []common.AdHocFilter{
				{
					Key:      "Country Name",
					Operator: "=",
					Value:    "A",
				},
			},
		},
	}

	ctx := context.Background()

	response, err := h.Endpoints().Query(ctx, query.Request{Args: args})
	require.NoError(t, err)
	assert.Equal(t, &query.TableResponse{Columns: []query.Column{
		{Text: "timestamp", Data: query.TimeColumn{time.Date(2020, time.November, 1, 0, 0, 0, 0, time.UTC), time.Date(2020, time.November, 2, 0, 0, 0, 0, time.UTC)}},
		{Text: "confirmed", Data: query.NumberColumn{1, 3}},
		{Text: "deaths", Data: query.NumberColumn{0, 0}},
	}}, response)

	mock.AssertExpectationsForObjects(t, dbh)
}

func TestCumulativeHandler_Tags(t *testing.T) {
	dbh := &mockCovidStore.CovidStore{}
	h := summarized.CumulativeHandler{Retriever: summarized.Retriever{DB: dbh}}

	ctx := context.Background()

	keys := h.Endpoints().TagKeys(ctx)
	assert.Equal(t, []string{"Country Name"}, keys)

	dbh.On("GetAllCountryNames").Return(nil, errors.New("db error")).Once()
	_, err := h.Endpoints().TagValues(ctx, keys[0])
	require.Error(t, err)

	dbh.On("GetAllCountryNames").Return([]string{"A", "B"}, nil)
	var values []string
	values, err = h.Endpoints().TagValues(ctx, keys[0])
	require.NoError(t, err)
	assert.Equal(t, []string{"A", "B"}, values)

	mock.AssertExpectationsForObjects(t, dbh)
}

var (
	dbContents = []models.CountryEntry{
		{
			Timestamp: time.Date(2020, time.November, 1, 0, 0, 0, 0, time.UTC),
			Code:      "A",
			Name:      "A",
			Confirmed: 1,
			Recovered: 0,
			Deaths:    0,
		},
		{
			Timestamp: time.Date(2020, time.November, 2, 0, 0, 0, 0, time.UTC),
			Code:      "B",
			Name:      "B",
			Confirmed: 3,
			Recovered: 0,
			Deaths:    0,
		},
		{
			Timestamp: time.Date(2020, time.November, 2, 0, 0, 0, 0, time.UTC),
			Code:      "A",
			Name:      "A",
			Confirmed: 3,
			Recovered: 1,
			Deaths:    0,
		},
		{
			Timestamp: time.Date(2020, time.November, 4, 0, 0, 0, 0, time.UTC),
			Code:      "B",
			Name:      "B",
			Confirmed: 10,
			Recovered: 5,
			Deaths:    1,
		},
	}

	dbTotals = []models.CountryEntry{
		{
			Timestamp: time.Date(2020, time.November, 1, 0, 0, 0, 0, time.UTC),
			Confirmed: 1,
			Recovered: 0,
			Deaths:    0,
		},
		{
			Timestamp: time.Date(2020, time.November, 2, 0, 0, 0, 0, time.UTC),
			Confirmed: 3,
			Recovered: 0,
			Deaths:    0,
		},
		{
			Timestamp: time.Date(2020, time.November, 3, 0, 0, 0, 0, time.UTC),
			Confirmed: 3,
			Recovered: 1,
			Deaths:    0,
		},
		{
			Timestamp: time.Date(2020, time.November, 4, 0, 0, 0, 0, time.UTC),
			Confirmed: 10,
			Recovered: 5,
			Deaths:    1,
		},
	}
)

func filterByName(input []models.CountryEntry, name string) (output []models.CountryEntry) {
	for _, entry := range input {
		if entry.Name == name {
			output = append(output, entry)
		}
	}
	return
}
