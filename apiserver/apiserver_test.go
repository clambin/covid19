package apiserver

import(
	"time"
	"testing"
	"encoding/json"
	"net/http"
	"net/http/httptest"

	"github.com/stretchr/testify/assert"

	"covid19api/coviddb"
)

func TestServerHello(t *testing.T) {
	req, err := http.NewRequest("GET", "/", nil)
    if err != nil {
        t.Fatal(err)
    }
    recorder := httptest.NewRecorder()
    server   := Server(Handler(nil))
    handler := http.HandlerFunc(server.hello)

    handler.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
	expected := `Hello`
	assert.Equal(t, expected, recorder.Body.String())
}

func TestHandlerSearch(t *testing.T) {
	req, err := http.NewRequest("POST", "/search", nil)
	if err != nil {
		t.Fatal(err)
	}
	recorder := httptest.NewRecorder()
	server   := Server(Handler(nil))
	handler := http.HandlerFunc(server.search)

	handler.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
	expected := `["active","active-delta","confirmed","confirmed-delta","death","death-delta","recovered","recovered-delta"]`
	assert.Equal(t, expected, recorder.Body.String())
}

func TestIsValidTarget(t *testing.T) {
	var tests = []struct {
		input string
		expected bool
	}{
		{"active", true},
		{"active-delta", true},
		{"invalid", false},
		{"zzzzz", false},
	}

	for _, tt := range tests {
		assert.Equal(t, isValidTarget(tt.input, targets), tt.expected)
	}
}

func parseDate(dateString string) (time.Time) {
		date, _ := time.Parse("2006-01-02T15:04:05.000Z", dateString)
        return date
}

func TestBuildSeries (t *testing.T) {
	entries := []coviddb.CountryEntry{
		coviddb.CountryEntry{
			Timestamp: parseDate("2020-11-01T00:00:00.000Z"),
			Code: "BE",
			Name: "Belgium",
			Confirmed: 1,
			Recovered: 0,
			Deaths: 0},
		coviddb.CountryEntry{
			Timestamp: parseDate("2020-11-02T00:00:00.000Z"),
			Code: "US",
			Name: "United States",
			Confirmed: 3,
			Recovered: 0,
			Deaths: 0},
		coviddb.CountryEntry{
			Timestamp: parseDate("2020-11-02T00:00:00.000Z"),
			Code: "BE",
			Name: "Belgium",
			Confirmed: 3,
			Recovered: 1,
			Deaths: 0},
		coviddb.CountryEntry{
			Timestamp: parseDate("2020-11-04T00:00:00.000Z"),
			Code: "US",
			Name: "United States",
			Confirmed: 10,
			Recovered: 5,
			Deaths: 1}}

	series := buildSeries(entries, []string{"confirmed", "confirmed-delta"})

	assert.Equal(t, 2,                 len(series))
	assert.Equal(t, "confirmed",       series[0].Target)
	assert.Equal(t, int64(1),          series[0].Datapoints[0][0])
	assert.Equal(t, int64(6),          series[0].Datapoints[1][0])
	assert.Equal(t, int64(13),         series[0].Datapoints[2][0])
	assert.Equal(t, "confirmed-delta", series[1].Target)
	assert.Equal(t, int64(1),          series[1].Datapoints[0][0])
	assert.Equal(t, int64(5),          series[1].Datapoints[1][0])
	assert.Equal(t, int64(7),          series[1].Datapoints[2][0])

	text, _ := json.Marshal(series)

	assert.Equal(t, "[{\"target\":\"confirmed\",\"datapoints\":[[1,1604188800000],[6,1604275200000],[13,1604448000000]]},{\"target\":\"confirmed-delta\",\"datapoints\":[[1,1604188800000],[5,1604275200000],[7,1604448000000]]}]", string(text))
}


