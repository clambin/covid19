package apiserver

import(
	"time"
	"testing"
	"net/http"
	"net/http/httptest"
	"bytes"

	"github.com/stretchr/testify/assert"
	log     "github.com/sirupsen/logrus"
)

func TestHello(t *testing.T) {
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}
	server   := Create(newAPIHandler(), -1)
	handler  := http.HandlerFunc(server.hello)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
	expected := `Hello`
	assert.Equal(t, expected, recorder.Body.String())
}

func TestSearch(t *testing.T) {
	req, err := http.NewRequest("POST", "/search", nil)
	if err != nil {
		t.Fatal(err)
	}
	server   := Create(newAPIHandler(), -1)
	handler  := http.HandlerFunc(server.search)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
	expected := `["A","B"]`
	assert.Equal(t, expected, recorder.Body.String())
}

func TestParseRequest(t *testing.T) {
	body := []byte(`
		{
			"app": "dashboard",
			"requestId": "Q111",
			"timezone": "browser",
			"panelId": 23763571993,
			"dashboardId": 160,
	 		"range": { 
	 			"from": "2019-12-31T23:59:59.000Z", 
				"to": "2020-12-31T23:59:59.000Z",
				"raw": {
					"from": "now-1y",
					"to": "now"
				}
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
			],
			"maxDataPoints":991,
			"scopedVars":{
				"__interval":   { "text":"6h",       "value":"6h" },
				"__interval_ms":{ "text":"21600000", "value":21600000 }
			},
			"startTime":1607274352883,
			"rangeRaw":{
				"from":"now-1y",
				"to":"now"
			},
			"adhocFilters":[]
		}
	`)
	targets := []string{"confirmed", "recovered", "death", "active"}
	reader := bytes.NewReader(body)
	params, err := parseRequest(reader, targets)

	assert.Nil(t, err)
	assert.True(t, time.Date(2019, time.December,  31, 23, 59, 59, 0, time.UTC).Equal(params.Range.From))
	assert.True(t, time.Date(2020, time.December,  31, 23, 59, 59, 0, time.UTC).Equal(params.Range.To))
	if err == nil {
		log.Printf("%v", params.Targets)
	}
	for _, target := range targets {
		found := false
		for _, parsedTarget := range params.Targets {
			if parsedTarget.Target == target {
				found = true
				break
			}
		}
		assert.True(t, found, target)
	}

}

func TestQuery(t *testing.T) {
	body := []byte(`{
			"range": { 
				"from": "2020-01-01T00:00:00.000Z", 
				"to": "2020-12-31T23:59:59.999Z"
			},
			"targets": [
				{ "target": "A" },
				{ "target": "B" },
				{ "target": "C" }
			]}`)
	req, err := http.NewRequest("POST", "/query", bytes.NewBuffer(body))
	if err != nil {
		t.Fatal(err)
	}

	server   := Create(newAPIHandler(), -1)
	handler  := http.HandlerFunc(server.query)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, req)
	assert.Equal(t, http.StatusOK, recorder.Code)
	expected := `{"target":"","datapoints":null}`
	assert.Equal(t, expected, recorder.Body.String())
}

func TestQueryBadRequest(t *testing.T) {
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
	server   := Create(newAPIHandler(), -1)
	handler  := http.HandlerFunc(server.query)

	handler.ServeHTTP(recorder, req)
	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

//
// Test APIHandler
//

type testAPIHandler struct {
}

func newAPIHandler() (*testAPIHandler) {
	return &testAPIHandler{}
}

var (
	targets = []string{"A", "B"}
)

func (apihandler *testAPIHandler) Search() ([]string) {
	return targets
}

func (apihandler *testAPIHandler) Query(params *APIQueryRequest) (*APIQueryResponse, error) {
	var response = APIQueryResponse{}
	return &response, nil
}


