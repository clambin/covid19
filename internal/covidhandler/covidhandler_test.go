package covidhandler_test

import (
	"github.com/clambin/covid19/internal/covidcache"
	"github.com/clambin/covid19/internal/coviddb"
	mockdb "github.com/clambin/covid19/internal/coviddb/mock"
	"github.com/clambin/covid19/internal/covidhandler"
	grafana_json "github.com/clambin/grafana-json"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestHandlerHandler(t *testing.T) {
	dbh := mockdb.Create([]coviddb.CountryEntry{
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
	})
	cache := covidcache.New(dbh)
	go cache.Run()

	handler, _ := covidhandler.Create(cache)

	// targets

	// Test Search
	targets := handler.Search()
	assert.Equal(t, covidhandler.Targets, targets)

	// Test Query
	request := grafana_json.QueryRequest{
		Range: grafana_json.QueryRequestRange{
			From: time.Now(),
			To:   time.Now(),
		},
		Targets: []grafana_json.QueryRequestTarget{
			{Target: "confirmed"},
			{Target: "confirmed-delta"},
			{Target: "death"},
			{Target: "death-delta"},
			{Target: "recovered"},
			{Target: "recovered-delta"},
			{Target: "active"},
			{Target: "active-delta"},
			{Target: "invalid"},
		},
	}

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
		responses, err := handler.Query(target, &request)

		if assert.Nil(t, err) {
			assert.Equal(t, len(testCase), len(responses.DataPoints))

			for index, entry := range testCase {
				assert.Equal(t, entry, responses.DataPoints[index].Value)
			}
		}

	}
}

func TestNoDB(t *testing.T) {
	_, err := covidhandler.Create(nil)

	assert.NotNil(t, err)
}

func BenchmarkHandlerQuery(b *testing.B) {
	// Build a large DB
	countries := []struct{ code, name string }{
		{code: "BE", name: "Belgium"},
		{code: "US", name: "USA"},
		{code: "FR", name: "France"},
		{code: "NL", name: "Netherlands"},
		{code: "UK", name: "United Kingdom"}}
	timestamp := time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC)
	entries := make([]coviddb.CountryEntry, 0)
	for i := 0; i < 365; i++ {
		for _, country := range countries {
			entries = append(entries, coviddb.CountryEntry{
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
	cache := covidcache.New(mockdb.Create(entries))
	handler, _ := covidhandler.Create(cache)

	request := grafana_json.QueryRequest{
		Range: grafana_json.QueryRequestRange{
			From: time.Now(),
			To:   time.Now(),
		},
		Targets: []grafana_json.QueryRequestTarget{
			{Target: "confirmed"},
			{Target: "confirmed-delta"},
			{Target: "death"},
			{Target: "death-delta"},
			{Target: "recovered"},
			{Target: "recovered-delta"},
			{Target: "active"},
			{Target: "active-delta"},
			{Target: "invalid"},
		},
	}

	b.ResetTimer()

	// Run the benchmark
	go cache.Run()
	for i := 0; i < 10; i++ {
		for _, target := range covidhandler.Targets {
			_, err := handler.Query(target, &request)
			assert.Nil(b, err)
		}
	}
}
