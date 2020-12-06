package covid_test

import (
	"time"
	"net/http"
	"bytes"
	"io/ioutil"

	"covid19/internal/covid"
)

var (
	testDBData = []covid.CountryEntry{
	covid.CountryEntry{
		Timestamp: parseDate("2020-11-01"),
		Code: "A",
		Name: "A",
		Confirmed: 1,
		Recovered: 0,
		Deaths: 0},
	covid.CountryEntry{
		Timestamp: parseDate("2020-11-02"),
		Code: "B",
		Name: "B",
		Confirmed: 3,
		Recovered: 0,
		Deaths: 0},
	covid.CountryEntry{
		Timestamp: parseDate("2020-11-02"),
		Code: "A",
		Name: "A",
		Confirmed: 3,
		Recovered: 1,
		Deaths: 0},
	covid.CountryEntry{
		Timestamp: parseDate("2020-11-04"),
		Code: "B",
		Name: "B",
		Confirmed: 10,
		Recovered: 5,
		Deaths: 1}}

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
	}`)

	lastChecked = time.Date(2020, time.December, 3, 11, 23, 52, 193000000, time.UTC)
	lastUpdate  = time.Date(2020, time.December, 3,  5, 28, 22,         0, time.UTC)
)

// parseDate helper function
func parseDate(dateString string) (time.Time) {
	date, _ := time.Parse("2006-01-02", dateString)
	return date
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
		Transport: RoundTripFunc(fn),
	}
}

// makeClient returns a stubbed covid.APIClient
func makeClient() (*covid.APIClient) {
	client := NewTestClient(func(req *http.Request) *http.Response {
		return &http.Response{
			StatusCode: 200,
			Header:	 make(http.Header),
			Body:       ioutil.NopCloser(bytes.NewBufferString(goodResponse)),
		}
	})

	return covid.NewAPIClient(client, "")
}

