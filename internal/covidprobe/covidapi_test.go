package covidprobe

import (
	"time"
	"bytes"
	"io/ioutil"
	"net/http"

	"testing"
	"github.com/stretchr/testify/assert"
)

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
		Transport: RoundTripFunc(fn),
	}
}

const (
	goodResponse = string(`
	{
		"error": false,
		"statusCode": 200,
		"message": "OK",
		"data": {
			"lastChecked": "2020-12-03T11:23:52.193Z",
			"covid19Stats": [
				{
					"city": null,
					"province": null,
					"country": "A",
					"lastUpdate": "2020-12-03T05:28:22+00:00",
					"keyId": "A",
					"confirmed": 3,
					"deaths": 2,
					"recovered": 1
				},
				{
					"city": null,
					"province": null,
					"country": "B",
					"lastUpdate": "2020-12-03T05:28:22+00:00",
					"keyId": "B",
					"confirmed": 6,
					"deaths": 5,
					"recovered": 4
				}
			]
		}
	}`)
)

var (
	lastChecked = time.Date(2020, time.December, 3, 11, 23, 52, 193000000, time.UTC)
	lastUpdate  = time.Date(2020, time.December, 3,  5, 28, 22,         0, time.UTC)
)

func TestRealClient(t *testing.T) {
	client := NewTestClient(func(req *http.Request) *http.Response {
	// Test request parameters
	// equals(t, req.URL.String(), "http://example.com/some/path")
		return &http.Response{
			StatusCode: 200,
			// Send response to be tested
			Body:       ioutil.NopCloser(bytes.NewBufferString(goodResponse)),
			// Must be set to non-nil value or it panics
			Header:     make(http.Header),
		}
	})

	apiClient := NewCovidAPIClient(client, "75f069b0d0mshb3c1017c5d7ee9dp11bc14jsn120bffb267a8")

	response, err := apiClient.GetStats()

	assert.Equal(t, nil,         err)
	assert.Equal(t, false,       response.Error)
	assert.Equal(t, 200,         response.StatusCode)
	assert.Equal(t, lastChecked, response.Data.LastChecked)
	assert.Equal(t, 2,           len(response.Data.Covid19Stats))
	assert.Equal(t, lastUpdate , response.Data.Covid19Stats[0].LastUpdate.UTC())
	assert.Equal(t, "A",         response.Data.Covid19Stats[0].Country)
	assert.Equal(t, 3,           response.Data.Covid19Stats[0].Confirmed)
	assert.Equal(t, 2,           response.Data.Covid19Stats[0].Deaths)
	assert.Equal(t, 1,           response.Data.Covid19Stats[0].Recovered)
	assert.Equal(t, lastUpdate , response.Data.Covid19Stats[1].LastUpdate.UTC())
	assert.Equal(t, "B",         response.Data.Covid19Stats[1].Country)
	assert.Equal(t, 6,           response.Data.Covid19Stats[1].Confirmed)
	assert.Equal(t, 5,           response.Data.Covid19Stats[1].Deaths)
	assert.Equal(t, 4,           response.Data.Covid19Stats[1].Recovered)
}
