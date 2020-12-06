package population

import (
	"net/http"
	"bytes"
	"io/ioutil"

	// "covid19/internal/coviddb"

)

var (
//	testDBData = []coviddb.CountryEntry{
//	coviddb.CountryEntry{
//		Timestamp: parseDate("2020-11-01"),
//		Code: "A",
//		Name: "A",
//		Confirmed: 1,
//		Recovered: 0,
//		Deaths: 0},
//	coviddb.CountryEntry{
//		Timestamp: parseDate("2020-11-02"),
//		Code: "B",
//		Name: "B",
//		Confirmed: 3,
//		Recovered: 0,
//		Deaths: 0},
//	coviddb.CountryEntry{
//		Timestamp: parseDate("2020-11-02"),
//		Code: "A",
//		Name: "A",
//		Confirmed: 3,
//		Recovered: 1,
//		Deaths: 0},
//	coviddb.CountryEntry{
//		Timestamp: parseDate("2020-11-04"),
//		Code: "B",
//		Name: "B",
//		Confirmed: 10,
//		Recovered: 5,
//		Deaths: 1}}
//
	goodResponse = string(`
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
	}`)
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

// makeClient returns a stubbed CovidAPIClient
func makeClient() (*APIClient) {
	client := NewTestClient(func(req *http.Request) *http.Response {
		return &http.Response{
			StatusCode: 200,
			Header:	 make(http.Header),
			Body:       ioutil.NopCloser(bytes.NewBufferString(goodResponse)),
		}
	})

	return NewAPIClient(client, "")
}

