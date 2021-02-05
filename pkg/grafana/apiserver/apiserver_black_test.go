package apiserver_test

import (
	"bytes"
	"errors"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"covid19/pkg/grafana/apiserver"
)

func TestAPIServer_Full(t *testing.T) {
	server := apiserver.Create(newAPIHandler(), 8080)

	go func() {
		err := server.Run()

		assert.Nil(t, err)
	}()

	time.Sleep(1 * time.Second)

	body, err := call("http://localhost:8080/", "GET", "")
	if assert.Nil(t, err) {
		assert.Equal(t, "Hello", body)
	}

	body, err = call("http://localhost:8080/metrics", "GET", "")
	if assert.Nil(t, err) {
		assert.Contains(t, body, "grafana_api_duration_seconds")
		assert.Contains(t, body, "grafana_api_duration_seconds_sum")
		assert.Contains(t, body, "grafana_api_duration_seconds_count")
	}

	body, err = call("http://localhost:8080/search", "POST", "")
	if assert.Nil(t, err) {
		assert.Equal(t, `["A","B","Crash"]`, body)
	}

	req := `{
	"maxDataPoints": 100,
	"interval": "1y",
	"range": {
		"from": "2020-01-01T00:00:00.000Z",
		"to": "2020-12-31T00:00:00.000Z"
	},
	"targets": [
		{ "target": "A", "type": "foo" },
		{ "target": "B", "type": "foo" }
	]
}`
	body, err = call("http://localhost:8080/query", "POST", req)

	if assert.Nil(t, err) {
		assert.Equal(t, `[{"target":"A","datapoints":[[1,2],[3,4]]},{"target":"B","datapoints":[[5,6],[7,8]]}]`, body)
	}
}

func BenchmarkAPIServer(b *testing.B) {
	server := apiserver.Create(newAPIHandler(), 8080)

	go func() {
		err := server.Run()

		assert.Nil(b, err)
	}()

	time.Sleep(1 * time.Second)

	req := `{
	"maxDataPoints": 100,
	"interval": "1y",
	"range": {
		"from": "2020-01-01T00:00:00.000Z",
		"to": "2020-12-31T00:00:00.000Z"
	},
	"targets": [
		{ "target": "A", "type": "foo" },
		{ "target": "B", "type": "foo" }
	]
}`
	var body string
	var err error

	for i := 0; i < 10000; i++ {
		body, err = call("http://localhost:8080/query", "POST", req)
	}

	if assert.Nil(b, err) {
		assert.Equal(b, `[{"target":"A","datapoints":[[1,2],[3,4]]},{"target":"B","datapoints":[[5,6],[7,8]]}]`, body)
	}

}

func call(url, method, body string) (string, error) {
	client := &http.Client{}
	reqBody := bytes.NewBuffer([]byte(body))
	req, _ := http.NewRequest(method, url, reqBody)
	resp, err := client.Do(req)

	if err == nil {
		defer resp.Body.Close()
		if body, err := ioutil.ReadAll(resp.Body); err == nil {
			return string(body), nil
		}
	}

	return "", err
}

//
//
// Test APIHandler
//

type testAPIHandler struct {
}

func newAPIHandler() *testAPIHandler {
	return &testAPIHandler{}
}

var (
	targets = []string{"A", "B", "Crash"}
)

func (apiHandler *testAPIHandler) Search() []string {
	return targets
}

func (apiHandler *testAPIHandler) Query(request *apiserver.APIQueryRequest) ([]apiserver.APIQueryResponse, error) {
	var response = make([]apiserver.APIQueryResponse, 0)

	for _, target := range request.Targets {
		switch target.Target {
		case "A":
			response = append(response, apiserver.APIQueryResponse{
				Target:     "A",
				DataPoints: [][2]int64{{int64(1), int64(2)}, {int64(3), int64(4)}}})
		case "B":
			response = append(response, apiserver.APIQueryResponse{
				Target:     "B",
				DataPoints: [][2]int64{{int64(5), int64(6)}, {int64(7), int64(8)}}})
		case "Crash":
			return response, errors.New("server crash")
		}
	}

	return response, nil
}
