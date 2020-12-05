package apiserver

import(
	"time"
	"testing"
	"net/http"
	"net/http/httptest"
	"bytes"

	"github.com/stretchr/testify/assert"
	log     "github.com/sirupsen/logrus"

	"covid19/internal/coviddb"
	"covid19/internal/coviddb/mock"
)

func TestServerHello(t *testing.T) {
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}
	recorder := httptest.NewRecorder()
	server   := CreateGrafanaAPIServer(CreateCovidAPIHandler(nil), -1)
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
	server   := CreateGrafanaAPIServer(CreateCovidAPIHandler(nil), -1)
	handler := http.HandlerFunc(server.search)

	handler.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
	expected := `["active","active-delta","confirmed","confirmed-delta","death","death-delta","recovered","recovered-delta"]`
	assert.Equal(t, expected, recorder.Body.String())
}

func TestParseRequest(t *testing.T) {
	// validTargets := []string{"confirmed", "confirmed-delta"}
	body := []byte(`{
	 		"range": { 
	 			"from": "2020-01-01T00:00:00.000Z", 
				"to": "2020-12-31T23:59:59.000Z"
			},
			"targets": [
				{ "target": "confirmed" },
				{ "target": "confirmed-delta" },
				{ "target": "recovered" },
				{ "target": "recovered-delta" },
				{ "target": "death" },
				{ "target": "death-delta" },
				{ "target": "active" },
				{ "target": "active-delta" },
				{ "target": "invalid" }
			]}`)
	reader := bytes.NewReader(body)
	params, err := parseRequest(reader, targets)

	assert.Equal(t, nil,                        err)
	assert.Equal(t, time.Date(2020, time.January,   1,  0,  0,  0, 0, time.UTC), params.Range["from"])
	assert.Equal(t, time.Date(2020, time.December, 31, 23, 59, 59, 0, time.UTC), params.Range["to"])
	if err == nil {
		log.Printf("%v", params.Targets)
	}
}

func TestHandlerQuery(t *testing.T) {
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

	db := mock.Create(entries)

	body := []byte(`{
			"range": { 
				"from": "2020-01-01T00:00:00.000Z", 
				"to": "2020-12-31T23:59:59.999Z"
			},
			"targets": [
				{ "target": "confirmed" },
				{ "target": "confirmed-delta" },
				{ "target": "recovered" },
				{ "target": "recovered-delta" },
				{ "target": "death" },
				{ "target": "death-delta" },
				{ "target": "active" },
				{ "target": "active-delta" },
				{ "target": "invalid" }
			]}`)
	req, err := http.NewRequest("POST", "/query", bytes.NewBuffer(body))
	if err != nil {
		t.Fatal(err)
	}

	// log.SetLevel(log.DebugLevel)

	recorder := httptest.NewRecorder()
	server   := CreateGrafanaAPIServer(CreateCovidAPIHandler(db), -1)
	handler  := http.HandlerFunc(server.query)

	handler.ServeHTTP(recorder, req)
	assert.Equal(t, http.StatusOK, recorder.Code)
	expected := "[{\"target\":\"confirmed\",\"datapoints\":[[1,1604188800000],[6,1604275200000],[13,1604448000000]]},{\"target\":\"confirmed-delta\",\"datapoints\":[[1,1604188800000],[5,1604275200000],[7,1604448000000]]},{\"target\":\"recovered\",\"datapoints\":[[0,1604188800000],[1,1604275200000],[6,1604448000000]]},{\"target\":\"recovered-delta\",\"datapoints\":[[0,1604188800000],[1,1604275200000],[5,1604448000000]]},{\"target\":\"death\",\"datapoints\":[[0,1604188800000],[0,1604275200000],[1,1604448000000]]},{\"target\":\"death-delta\",\"datapoints\":[[0,1604188800000],[0,1604275200000],[1,1604448000000]]},{\"target\":\"active\",\"datapoints\":[[1,1604188800000],[5,1604275200000],[6,1604448000000]]},{\"target\":\"active-delta\",\"datapoints\":[[1,1604188800000],[4,1604275200000],[1,1604448000000]]}]"
	assert.Equal(t, expected, recorder.Body.String())
}

func TestHandlerQueryBadRequest(t *testing.T) {
	entries := []coviddb.CountryEntry{}
	db := mock.Create(entries)

	body := []byte(`{
			"range": { 
				"from": "2020-01-01T00:00:00.000Z", 
				"to": "notatimestamp"
			}}`)
	req, err := http.NewRequest("POST", "/query", bytes.NewBuffer(body))
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()
	server   := CreateGrafanaAPIServer(CreateCovidAPIHandler(db), -1)
	handler  := http.HandlerFunc(server.query)

	handler.ServeHTTP(recorder, req)
	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestHandlerQueryRequestError(t *testing.T) {
	body := []byte(`{
			"range": { 
				"from": "2020-01-01T00:00:00.000Z", 
				"to": "2020-12-31T23:59:59.999Z"
			},
			"targets": [
				{ "target": "confirmed" }
			]}`)
	req, err := http.NewRequest("POST", "/query", bytes.NewBuffer(body))
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()
	server   := CreateGrafanaAPIServer(CreateCovidAPIHandler(nil), -1)
	handler  := http.HandlerFunc(server.query)

	handler.ServeHTTP(recorder, req)
	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
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
	entries := make([]coviddb.CountryEntry, 0)
	for i:=0; i<365; i++ {
		for _, country := range countries {
				entries = append(entries, coviddb.CountryEntry{Timestamp: timestamp, Code: country.code, Name: country.name, Confirmed: int64(i), Recovered: 0, Deaths: 0})
		}
		timestamp = timestamp.Add(24 * time.Hour)
	}
	db := mock.Create(entries)

	// Set up client & server
	body := []byte(`{
			"range": { 
				"from": "2020-01-01T00:00:00.000Z", 
				"to": "2020-12-31T23:59:59.999Z"
			},
			"targets": [
				{ "target": "confirmed" },
				{ "target": "confirmed-delta" },
				{ "target": "recovered" },
				{ "target": "recovered-delta" },
				{ "target": "death" },
				{ "target": "death-delta" },
				{ "target": "active" },
				{ "target": "active-delta" }
			]}`)
	req, err := http.NewRequest("POST", "/query", bytes.NewBuffer(body))
	if err != nil {
		b.Fatal(err)
	}
	recorder := httptest.NewRecorder()
	server   := CreateGrafanaAPIServer(CreateCovidAPIHandler(db), -1)
	handler  := http.HandlerFunc(server.query)

	// Run the benchmark
	b.ResetTimer()
	handler.ServeHTTP(recorder, req)
	assert.Equal(b, http.StatusOK, recorder.Code)
}
