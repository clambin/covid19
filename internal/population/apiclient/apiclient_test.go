package apiclient_test

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/clambin/httpstub"
	"github.com/stretchr/testify/assert"

	"covid19/internal/population/apiclient"
)

func TestGetPopulation(t *testing.T) {
	apiClient := apiclient.NewWithHTTPClient(httpstub.NewTestClient(loopback), "")

	response, err := apiClient.GetPopulation()

	assert.Equal(t, nil, err)
	assert.Equal(t, 2, len(response))
	population, ok := response["BE"]
	assert.Equal(t, true, ok)
	assert.Equal(t, int64(11248330), population)
	population, ok = response["US"]
	assert.Equal(t, true, ok)
	assert.Equal(t, int64(321645000), population)
	_, ok = response["??"]
	assert.Equal(t, false, ok)
}

// Loopback function

// makeClient returns a stubbed CovidAPIClient
func loopback(_ *http.Request) *http.Response {
	return &http.Response{
		StatusCode: 200,
		Header:     make(http.Header),
		Body:       ioutil.NopCloser(bytes.NewBufferString(goodResponse)),
	}
}

var goodResponse = `
	{
		"data": {
			"countries": [
				{
    				"name": "Belgium",
					"countryCode": "BE",
					"population": "11248330"
				},
				{
					"name": "United States",
					"countryCode": "US", 
					"population": "321645000"
				}
			]
		}
	}`
