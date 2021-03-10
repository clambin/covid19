package covidprobe_test

import (
	"bytes"
	"github.com/clambin/covid19/internal/covidprobe"
	"github.com/clambin/gotools/httpstub"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"testing"
)

func TestGetCountryStats(t *testing.T) {
	apiClient := covidprobe.NewAPIClient("1234")
	apiClient.(*covidprobe.RapidAPIClient).Client.Client = httpstub.NewTestClient(loopback)

	response, err := apiClient.GetCountryStats()

	assert.Equal(t, nil, err)
	assert.Equal(t, 2, len(response))
	stats, ok := response["A"]
	assert.Equal(t, true, ok)
	assert.Equal(t, int64(3), stats.Confirmed)
	assert.Equal(t, int64(2), stats.Deaths)
	assert.Equal(t, int64(1), stats.Recovered)
	stats, ok = response["B"]
	assert.Equal(t, true, ok)
	assert.Equal(t, int64(6), stats.Confirmed)
	assert.Equal(t, int64(5), stats.Deaths)
	assert.Equal(t, int64(4), stats.Recovered)
}

// loopback function
func loopback(_ *http.Request) *http.Response {
	return &http.Response{
		StatusCode: 200,
		Header:     make(http.Header),
		Body:       ioutil.NopCloser(bytes.NewBufferString(goodResponse)),
	}
}

const goodResponse = `
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
					"city": "B.1",
					"province": null,
					"country": "B",
					"lastUpdate": "2020-12-03T05:28:22+00:00",
					"keyId": "B",
					"confirmed": 5,
					"deaths": 4,
					"recovered": 3
				},
				{
					"city": "B.2",
					"province": null,
					"country": "B",
					"lastUpdate": "2020-12-03T05:28:22+00:00",
					"keyId": "B",
					"confirmed": 1,
					"deaths": 1,
					"recovered": 1
				}
			]
		}
	}`
