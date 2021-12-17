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
	"testing"
	"time"
)

func BenchmarkHandlerTableQuery(b *testing.B) {
	// Build a large PopulationDB
	countries := []struct{ code, name string }{
		{code: "BE", name: "Belgium"},
		{code: "US", name: "USA"},
		{code: "FR", name: "France"},
		{code: "NL", name: "Netherlands"},
		{code: "UK", name: "United Kingdom"}}
	timestamp := time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC)
	entries := make([]models.CountryEntry, 0)
	for i := 0; i < 365; i++ {
		for _, country := range countries {
			entries = append(entries, models.CountryEntry{
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
	for i := 0; i < 100; i++ {
		for _, target := range []string{"incremental", "cumulative"} {
			_, err := h.Endpoints().TableQuery(ctx, target, &tableArgs)
			assert.NoError(b, err)
		}
	}

	mock.AssertExpectationsForObjects(b, dbh)
}
