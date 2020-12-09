package covidhandler

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	// log     "github.com/sirupsen/logrus"

	"covid19/internal/coviddb"
	"covid19/internal/coviddb/testdb"
	"covid19/pkg/grafana/apiserver"
)

func TestHandlerHandler(t *testing.T) {
	db := testdb.CreateWithData()
	handler, _ := Create(db)

	// Test Search
	targets := handler.Search()
	assert.Equal(t, []string{"active", "active-delta", "confirmed", "confirmed-delta", "death", "death-delta", "recovered", "recovered-delta"}, targets)

	// Test Query
	request := apiserver.APIQueryRequest{
		Range: struct {
			From time.Time
			To   time.Time
		}{
			From: time.Now(),
			To:   time.Now()},
		Targets: []struct{ Target string }{
			{Target: "confirmed"},
			{Target: "confirmed-delta"},
			{Target: "death"},
			{Target: "death-delta"},
			{Target: "recovered"},
			{Target: "recovered-delta"},
			{Target: "active"},
			{Target: "active-delta"},
			{Target: "invalid"},
		}}

	testCases := map[string][][2]int64{
		"confirmed":       {[2]int64{1, 1604188800000}, [2]int64{6, 1604275200000}, [2]int64{13, 1604448000000}},
		"confirmed-delta": {[2]int64{1, 1604188800000}, [2]int64{5, 1604275200000}, [2]int64{7, 1604448000000}},
		"death":           {[2]int64{0, 1604188800000}, [2]int64{0, 1604275200000}, [2]int64{1, 1604448000000}},
		"death-delta":     {[2]int64{0, 1604188800000}, [2]int64{0, 1604275200000}, [2]int64{1, 1604448000000}},
		"recovered":       {[2]int64{0, 1604188800000}, [2]int64{1, 1604275200000}, [2]int64{6, 1604448000000}},
		"recovered-delta": {[2]int64{0, 1604188800000}, [2]int64{1, 1604275200000}, [2]int64{5, 1604448000000}},
		"active":          {[2]int64{1, 1604188800000}, [2]int64{5, 1604275200000}, [2]int64{6, 1604448000000}},
		"active-delta":    {[2]int64{1, 1604188800000}, [2]int64{4, 1604275200000}, [2]int64{1, 1604448000000}},
	}

	responses, err := handler.Query(&request)
	assert.Nil(t, err)
	assert.Equal(t, len(testCases), len(responses))

	incides := make(map[string]int, 0)
	for index, response := range responses {
		incides[response.Target] = index
	}
	assert.Equal(t, len(responses), len(incides))

	for target, expected := range testCases {
		index, ok := incides[target]
		assert.True(t, ok)
		assert.Equal(t, target, responses[index].Target)
		assert.Equal(t, expected, responses[index].Datapoints, target)
	}
}

func TestNoDB(t *testing.T) {
	_, err := Create(nil)

	assert.NotNil(t, err)
}

func BenchmarkHandlerQuery(b *testing.B) {
	// Build a large DB
	type country struct{ code, name string }
	countries := []country{
		{code: "BE", name: "Belgium"},
		{code: "US", name: "USA"},
		{code: "FR", name: "France"},
		{code: "NL", name: "Netherlands"},
		{code: "UK", name: "United Kingdom"}}
	timestamp := time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC)
	entries := make([]coviddb.CountryEntry, 0)
	for i := 0; i < 365; i++ {
		for _, country := range countries {
			entries = append(entries, coviddb.CountryEntry{Timestamp: timestamp, Code: country.code, Name: country.name, Confirmed: int64(i), Recovered: 0, Deaths: 0})
		}
		timestamp = timestamp.Add(24 * time.Hour)
	}
	db := testdb.Create(entries)
	handler, _ := Create(db)

	request := apiserver.APIQueryRequest{
		Range: struct {
			From time.Time
			To   time.Time
		}{
			From: time.Now(),
			To:   time.Now()},
		Targets: []struct{ Target string }{
			{Target: "confirmed"},
			{Target: "confirmed-delta"},
			{Target: "recovered"},
			{Target: "recovered-delta"},
			{Target: "death"},
			{Target: "death-delta"},
			{Target: "active"},
			{Target: "active-delta"},
		}}

	// Run the benchmark
	b.ResetTimer()
	_, err := handler.Query(&request)
	assert.Nil(b, err)
}
