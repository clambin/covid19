package probe_test

import (
	"bytes"
	"fmt"
	"github.com/clambin/covid19/population/probe"
	"github.com/clambin/gotools/httpstub"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"testing"
)

func TestGetPopulation(t *testing.T) {
	apiClient := probe.RapidAPIClient{
		HTTPClient: httpstub.NewTestClient(serverStub),
		APIKey:     "1234",
	}

	population, err := apiClient.GetPopulation("Belgium")

	assert.NoError(t, err)
	assert.Equal(t, int64(20), population)

	population, err = apiClient.GetPopulation("United States")

	assert.NoError(t, err)
	assert.Equal(t, int64(40), population)

	population, err = apiClient.GetPopulation("??")

	assert.Error(t, err)
}

func serverStub(req *http.Request) *http.Response {
	var response string
	if req.URL.Path == "/population" {
		switch req.URL.RawQuery {
		case "country_name=Belgium":
			response = fmt.Sprintf(goodResponse, "Belgium", 20)
		case "country_name=United+States":
			response = fmt.Sprintf(goodResponse, "United States", 40)
		case "country_name=Faroe+Islands":
			response = fmt.Sprintf(goodResponse, "Faroe Islands", 5)
		}
	}

	if response == "" {
		return &http.Response{StatusCode: http.StatusNotFound}
	}

	return &http.Response{
		StatusCode: 200,
		Header:     make(http.Header),
		Body:       ioutil.NopCloser(bytes.NewBufferString(response)),
	}
}

var goodResponse = `
	{
		"ok": true,
		"body": {
			"country_name": "%s",
			"population": %d
		}
	}
`
