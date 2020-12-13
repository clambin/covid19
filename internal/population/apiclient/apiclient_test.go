package apiclient_test

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"covid19/internal/population/apiclient"
)

func TestGetPopulation(t *testing.T) {
	apiClient := makeClient()

	response, err := apiClient.GetPopulation()

	assert.Equal(t, nil, err)
	assert.Equal(t, 2, len(response))
	population, ok := response["BE"]
	assert.Equal(t, true, ok)
	assert.Equal(t, int64(11248330), population)
	population, ok = response["US"]
	assert.Equal(t, true, ok)
	assert.Equal(t, int64(321645000), population)
	population, ok = response["??"]
	assert.Equal(t, false, ok)
}

// Stubbing the API Call

// RoundTripFunc .
type RoundTripFunc func(req *http.Request) *http.Response

// RoundTrip .
func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

//NewTestClient returns *http.Client with Transport replaced to avoid making real calls
func NewTestClient(fn RoundTripFunc) *http.Client {
	return &http.Client{
		Transport: fn,
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

// makeClient returns a stubbed CovidAPIClient
func makeClient() *apiclient.APIClient {
	client := NewTestClient(func(req *http.Request) *http.Response {
		return &http.Response{
			StatusCode: 200,
			Header:     make(http.Header),
			Body:       ioutil.NopCloser(bytes.NewBufferString(goodResponse)),
		}
	})

	return apiclient.NewWithHTTPClient(client, "")
}
