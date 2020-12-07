package covidhandler

import(
	"time"
	"testing"

	"github.com/stretchr/testify/assert"
	// log     "github.com/sirupsen/logrus"

	"covid19/pkg/grafana/apiserver"
	"covid19/internal/covid"
	"covid19/internal/covid/mock"
)

func TestHandlerHandler(t *testing.T) {
	entries := []covid.CountryEntry{
		covid.CountryEntry{
			Timestamp: parseDate("2020-11-01T00:00:00.000Z"),
			Code: "BE",
			Name: "Belgium",
			Confirmed: 1,
			Recovered: 0,
			Deaths: 0},
		covid.CountryEntry{
			Timestamp: parseDate("2020-11-02T00:00:00.000Z"),
			Code: "US",
			Name: "United States",
			Confirmed: 3,
			Recovered: 0,
			Deaths: 0},
		covid.CountryEntry{
			Timestamp: parseDate("2020-11-02T00:00:00.000Z"),
			Code: "BE",
			Name: "Belgium",
			Confirmed: 3,
			Recovered: 1,
			Deaths: 0},
		covid.CountryEntry{
			Timestamp: parseDate("2020-11-04T00:00:00.000Z"),
			Code: "US",
			Name: "United States",
			Confirmed: 10,
			Recovered: 5,
			Deaths: 1}}

	db := mock.Create(entries)

	handler := Create(db)

	// Test Search
	targets := handler.Search()
	assert.Equal(t, []string([]string{"active", "active-delta", "confirmed", "confirmed-delta", "death", "death-delta", "recovered", "recovered-delta"}), targets)

	// Test Query
	request := apiserver.APIQueryRequest{
			Range: struct{From time.Time; To time.Time}{
				From: time.Now(),
				To: time.Now()},
			Targets: []struct{Target string}{
				struct{Target string}{ Target: "confirmed" },
				struct{Target string}{ Target: "confirmed-delta" },
				struct{Target string}{ Target: "death" },
				struct{Target string}{ Target: "death-delta" },
				struct{Target string}{ Target: "recovered" },
				struct{Target string}{ Target: "recovered-delta" },
				struct{Target string}{ Target: "active" },
				struct{Target string}{ Target: "active-delta" },
				struct{Target string}{ Target: "invalid" },
		}}

	testCases := map[string][][2]int64{
		"confirmed":       [][2]int64{[2]int64{1, 1604188800000}, [2]int64{6, 1604275200000}, [2]int64{13, 1604448000000}},
		"confirmed-delta": [][2]int64{[2]int64{1, 1604188800000}, [2]int64{5, 1604275200000}, [2]int64{7, 1604448000000}},
		"death":           [][2]int64{[2]int64{0, 1604188800000}, [2]int64{0, 1604275200000}, [2]int64{1, 1604448000000}},
		"death-delta":     [][2]int64{[2]int64{0, 1604188800000}, [2]int64{0, 1604275200000}, [2]int64{1, 1604448000000}},
		"recovered":       [][2]int64{[2]int64{0, 1604188800000}, [2]int64{1, 1604275200000}, [2]int64{6, 1604448000000}},
		"recovered-delta": [][2]int64{[2]int64{0, 1604188800000}, [2]int64{1, 1604275200000}, [2]int64{5, 1604448000000}},
		"active":          [][2]int64{[2]int64{1, 1604188800000}, [2]int64{5, 1604275200000}, [2]int64{6, 1604448000000}},
		"active-delta":    [][2]int64{[2]int64{1, 1604188800000}, [2]int64{4, 1604275200000}, [2]int64{1, 1604448000000}},

	}

	responses, err := handler.Query(&request)
	assert.Nil(t, err)
	assert.Equal(t, len(testCases), len(responses))

	incides := make(map[string]int, 0)
	for index, response := range responses{
		incides[response.Target] = index
	}
	assert.Equal(t, len(responses), len(incides))

	for target, expected := range testCases {
		index, ok := incides[target]
		assert.True(t, ok)
		assert.Equal(t, target,   responses[index].Target)
		assert.Equal(t, expected, responses[index].Datapoints, target)
	}
}

func TestNoDB(t *testing.T) {
	handler := Create(nil)

	request := apiserver.APIQueryRequest{
			Range: struct{From time.Time; To time.Time}{
				From: time.Now(),
				To: time.Now()},
			Targets: []struct{Target string}{
				struct{Target string}{ Target: "confirmed" }}}

	_, err := handler.Query(&request)
	assert.NotNil(t, err)
}

func parseDate(dateString string) (time.Time) {
		date, _ := time.Parse("2006-01-02T15:04:05.000Z", dateString)
		return date
}

func BenchmarkHandlerQuery(b *testing.B) {
	// Build a large DB
	type country struct{code, name string}
	countries := []country{
			country {code:"BE", name:"Belgium"},
			country {code:"US", name:"USA"},
			country {code:"FR", name:"France"},
			country {code:"NL", name:"Netherlands"},
			country {code:"UK", name:"United Kingdom"}}
	timestamp := time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC)
	entries := make([]covid.CountryEntry, 0)
	for i:=0; i<365; i++ {
		for _, country := range countries {
				entries = append(entries, covid.CountryEntry{Timestamp: timestamp, Code: country.code, Name: country.name, Confirmed: int64(i), Recovered: 0, Deaths: 0})
		}
		timestamp = timestamp.Add(24 * time.Hour)
	}
	db := mock.Create(entries)

	handler := Create(db)

	request := apiserver.APIQueryRequest{
			Range: struct{From time.Time; To time.Time}{
				From: time.Now(),
				To: time.Now()},
			Targets: []struct{Target string}{
				struct{Target string}{ Target: "confirmed" },
				struct{Target string}{ Target: "confirmed-delta" },
				struct{Target string}{ Target: "recovered" },
				struct{Target string}{ Target: "recovered-delta" },
				struct{Target string}{ Target: "death" },
				struct{Target string}{ Target: "death-delta" },
				struct{Target string}{ Target: "active" },
				struct{Target string}{ Target: "active-delta" },
		}}

	// Run the benchmark
	b.ResetTimer()
	_, err := handler.Query(&request)
	assert.Nil(b, err)
}
