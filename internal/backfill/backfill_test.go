package backfill

import (
	"time"
	"net/http"
	"io/ioutil"
	"bytes"

	"testing"
	"github.com/stretchr/testify/assert"

	"covid19/internal/covid"
	"covid19/internal/covid/mock"
)

func DontTestAPICall(t *testing.T) {
	backfiller := Create(nil)

	result, err := backfiller.getCountries()
	assert.Nil(t, err)
	assert.Equal(t, "BE", result["belgium"].Code)
	assert.Equal(t, "Belgium", result["belgium"].Name)

	entries, err := backfiller.getHistoricalData("belgium")
	assert.Nil(t, err)
	assert.Less(t, 0, len(entries))
	assert.Zero(t, entries[0].Confirmed)
	assert.Zero(t, entries[0].Deaths)
	assert.Zero(t, entries[0].Recovered)
}

func TestBackfiller(t *testing.T) {
	db := mock.Create([]covid.CountryEntry{
		covid.CountryEntry{
			Timestamp: time.Date(2019, time.December, 31, 23, 59, 59, 0, time.UTC),
			Code:      "BE",
			Name:      "Belgium",
			Confirmed: 0,
			Deaths:    0,
			Recovered: 0}})

	backfiller := CreateWithClient(db, makeHTTPClient())

	err := backfiller.Run()
	assert.Nil(t, err)

	records, err := db.List(time.Now())
	assert.Nil(t, err)
	assert.Equal(t, 3, len(records))
}

// Stubbing the API Call

var goodResponse = string(`[
  {
    "Country": "Belgium",
    "CountryCode": "",
    "Province": "",
    "City": "",
    "CityCode": "",
    "Lat": "0",
    "Lon": "0",
    "Confirmed": 0,
    "Deaths": 0,
    "Recovered": 0,
    "Active": 0,
    "Date": "2020-01-22T00:00:00Z"
  },
  {
    "Country": "Belgium",
    "CountryCode": "",
    "Province": "",
    "City": "",
    "CityCode": "",
    "Lat": "0",
    "Lon": "0",
    "Confirmed": 1,
    "Deaths": 0,
    "Recovered": 0,
    "Active": 1,
    "Date": "2020-02-04T00:00:00Z"
  }
]`)

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

// makeClient returns a stubbed httpClient
func makeHTTPClient() (*http.Client) {
	return NewTestClient(func(req *http.Request) *http.Response {
		return &http.Response{
			StatusCode: 200,
			Header:     make(http.Header),
			Body:       ioutil.NopCloser(bytes.NewBufferString(goodResponse)),
		}
    })
}
