package handler_test

import (
	"context"
	"github.com/clambin/covid19/cache"
	mockCovidStore "github.com/clambin/covid19/covid/store/mocks"
	"github.com/clambin/covid19/handler"
	"github.com/clambin/covid19/models"
	grafanaJson "github.com/clambin/grafana-json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	// "github.com/stretchr/testify/require"
	"testing"
	"time"
)

var dbContents = []*models.CountryEntry{
	{
		Timestamp: time.Date(2020, time.November, 1, 0, 0, 0, 0, time.UTC),
		Code:      "A",
		Name:      "A",
		Confirmed: 1,
		Recovered: 0,
		Deaths:    0},
	{
		Timestamp: time.Date(2020, time.November, 2, 0, 0, 0, 0, time.UTC),
		Code:      "B",
		Name:      "B",
		Confirmed: 3,
		Recovered: 0,
		Deaths:    0},
	{
		Timestamp: time.Date(2020, time.November, 2, 0, 0, 0, 0, time.UTC),
		Code:      "A",
		Name:      "A",
		Confirmed: 3,
		Recovered: 1,
		Deaths:    0},
	{
		Timestamp: time.Date(2020, time.November, 4, 0, 0, 0, 0, time.UTC),
		Code:      "B",
		Name:      "B",
		Confirmed: 10,
		Recovered: 5,
		Deaths:    1,
	},
}

func TestCovidHandler_Search(t *testing.T) {
	store := &mockCovidStore.CovidStore{}
	store.On("List").Return(dbContents, nil)

	c := &cache.Cache{DB: store, Retention: 20 * time.Minute}
	h := handler.Handler{Cache: c}
	targets := h.Endpoints().Search()
	assert.Equal(t, handler.Targets, targets)

}

func TestTimeSeriesHandler(t *testing.T) {
	dbh := &mockCovidStore.CovidStore{}
	dbh.On("GetAll").Return(dbContents, nil)

	c := &cache.Cache{DB: dbh, Retention: 20 * time.Minute}
	h := handler.Handler{Cache: c}

	args := grafanaJson.TimeSeriesQueryArgs{
		CommonQueryArgs: grafanaJson.CommonQueryArgs{
			Range: grafanaJson.QueryRequestRange{
				From: time.Now(),
				To:   time.Now(),
			},
		},
	}

	ctx := context.Background()
	assert.Eventually(t, func() bool {
		responses, err := h.Endpoints().Query(ctx, "confirmed", &args)
		return err == nil && len(responses.DataPoints) > 0
	}, 500*time.Millisecond, 50*time.Millisecond)

	testCases := map[string][]int64{
		"confirmed":       {1, 6, 13},
		"confirmed-delta": {1, 5, 7},
		"death":           {0, 0, 1},
		"death-delta":     {0, 0, 1},
		"recovered":       {0, 1, 6},
		"recovered-delta": {0, 1, 5},
		"active":          {1, 5, 6},
		"active-delta":    {1, 4, 1},
	}

	for target, testCase := range testCases {
		responses, err := h.Endpoints().Query(context.Background(), target, &args)

		require.NoError(t, err, target)
		require.Equal(t, len(testCase), len(responses.DataPoints), target)
		for index, entry := range testCase {
			assert.Equal(t, entry, responses.DataPoints[index].Value, target, index)
		}
	}

	mock.AssertExpectationsForObjects(t, dbh)
}

func TestTableHandler(t *testing.T) {
	dbh := &mockCovidStore.CovidStore{}
	dbh.On("GetAll").Return(dbContents, nil)

	c := &cache.Cache{DB: dbh, Retention: 20 * time.Minute}
	h := handler.Handler{Cache: c}

	args := grafanaJson.TableQueryArgs{
		CommonQueryArgs: grafanaJson.CommonQueryArgs{
			Range: grafanaJson.QueryRequestRange{
				From: time.Now(),
				To:   time.Now(),
			},
		},
	}

	ctx := context.Background()
	assert.Eventually(t, func() bool {
		responses, err := h.Endpoints().TableQuery(ctx, "daily", &args)
		return err == nil && len(responses.Columns) > 0 && len(responses.Columns[0].Data.(grafanaJson.TableQueryResponseTimeColumn)) > 0
	}, 500*time.Millisecond, 50*time.Millisecond)

	testCases := map[string]map[string][]float64{
		"daily": {
			"confirmed": {1, 5, 7},
			"recovered": {0, 1, 5},
			"deaths":    {0, 0, 1},
		},
		"cumulative": {
			"active":    {1, 5, 6},
			"recovered": {0, 1, 6},
			"deaths":    {0, 0, 1},
		},
	}

	for target, testCase := range testCases {
		responses, err := h.Endpoints().TableQuery(context.Background(), target, &args)

		if assert.NoError(t, err, target) == false {
			continue
		}

		if assert.Len(t, responses.Columns, 4, target) == false {
			continue
		}

		for _, column := range responses.Columns {
			if column.Text == "timestamp" {
				continue
			}

			expected, ok := testCase[column.Text]

			if assert.True(t, ok, column.Text) == false {
				continue
			}

			if assert.Equal(t, len(expected), len(column.Data.(grafanaJson.TableQueryResponseNumberColumn)), target, column.Text) {
				for index, value := range expected {
					assert.Equal(t, value, column.Data.(grafanaJson.TableQueryResponseNumberColumn)[index], target, column.Text, index)
				}
			}
		}
	}
	mock.AssertExpectationsForObjects(t, dbh)
}

func BenchmarkHandlerQuery(b *testing.B) {
	// Build a large PopulationDB
	countries := []struct{ code, name string }{
		{code: "BE", name: "Belgium"},
		{code: "US", name: "USA"},
		{code: "FR", name: "France"},
		{code: "NL", name: "Netherlands"},
		{code: "UK", name: "United Kingdom"}}
	timestamp := time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC)
	entries := make([]*models.CountryEntry, 0)
	for i := 0; i < 365; i++ {
		for _, country := range countries {
			entries = append(entries, &models.CountryEntry{
				Timestamp: timestamp,
				Code:      country.code,
				Name:      country.name,
				Confirmed: int64(i),
				Recovered: 0,
				Deaths:    0,
			})
		}
		timestamp = timestamp.Add(24 * time.Hour)
	}
	dbh := &mockCovidStore.CovidStore{}
	dbh.On("GetAll").Return(entries, nil)

	c := &cache.Cache{DB: dbh, Retention: 20 * time.Minute}
	h := handler.Handler{Cache: c}

	seriesArgs := grafanaJson.TimeSeriesQueryArgs{
		CommonQueryArgs: grafanaJson.CommonQueryArgs{
			Range: grafanaJson.QueryRequestRange{
				From: time.Now(),
				To:   time.Now(),
			},
		},
	}
	ctx := context.Background()

	b.ResetTimer()

	// Run the benchmark
	for _, target := range handler.Targets {
		_, err := h.Endpoints().Query(ctx, target, &seriesArgs)
		assert.NoError(b, err)
	}

	mock.AssertExpectationsForObjects(b, dbh)
}

func BenchmarkHandlerTableQuery(b *testing.B) {
	// Build a large PopulationDB
	countries := []struct{ code, name string }{
		{code: "BE", name: "Belgium"},
		{code: "US", name: "USA"},
		{code: "FR", name: "France"},
		{code: "NL", name: "Netherlands"},
		{code: "UK", name: "United Kingdom"}}
	timestamp := time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC)
	entries := make([]*models.CountryEntry, 0)
	for i := 0; i < 365; i++ {
		for _, country := range countries {
			entries = append(entries, &models.CountryEntry{
				Timestamp: timestamp,
				Code:      country.code,
				Name:      country.name,
				Confirmed: int64(i),
				Recovered: 0,
				Deaths:    0,
			})
		}
		timestamp = timestamp.Add(24 * time.Hour)
	}

	dbh := &mockCovidStore.CovidStore{}
	dbh.On("GetAll").Return(entries, nil)

	c := &cache.Cache{DB: dbh, Retention: 20 * time.Minute}
	h := handler.Handler{Cache: c}

	tableArgs := grafanaJson.TableQueryArgs{
		CommonQueryArgs: grafanaJson.CommonQueryArgs{
			Range: grafanaJson.QueryRequestRange{
				From: time.Now(),
				To:   time.Now(),
			},
		},
	}

	ctx := context.Background()
	b.ResetTimer()

	// Run the benchmark
	for _, target := range []string{"daily", "cumulative"} {
		_, err := h.Endpoints().TableQuery(ctx, target, &tableArgs)
		assert.NoError(b, err)
	}

	mock.AssertExpectationsForObjects(b, dbh)
}
