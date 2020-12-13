package backfill

import (
	"bytes"
	"io/ioutil"
	"math/rand"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"covid19/internal/covid/db"
	"covid19/internal/covid/db/mock"
	"testing"
)

func TestBackFiller(t *testing.T) {
	covidDB := mock.Create([]db.CountryEntry{
		{
			Timestamp: time.Date(2020, time.February, 1, 0, 0, 0, 0, time.UTC),
			Code:      "BE",
			Name:      "Belgium",
			Confirmed: 0,
			Deaths:    0,
			Recovered: 0,
		}})

	backFiller := CreateWithClient(covidDB, makeHTTPClient())

	err := backFiller.Run()
	assert.Nil(t, err)

	records, err := covidDB.List(time.Now())
	assert.Nil(t, err)
	log.Debug(records)
	assert.Len(t, records, 3)

	latest, err := covidDB.ListLatestByCountry()
	assert.Nil(t, err)
	assert.Len(t, latest, 2)
	timestamp, ok := latest["Belgium"]
	assert.True(t, ok)
	assert.Equal(t, time.Date(2020, time.February, 1, 0, 0, 0, 0, time.UTC), timestamp)
	timestamp, ok = latest["Burma"]
	assert.True(t, ok)
	assert.Equal(t, time.Date(2020, time.January, 31, 0, 0, 0, 0, time.UTC), timestamp)
}

// Stubbing the API Call

var goodResponse = map[string]string{
	"/countries": `[
		 {
			"Country": "Belgium",
			"Slug": "belgium",
			"ISO2": "BE"
		},
		{
			"Country": "Myanmar",
			"Slug": "myanmar",
			"ISO2": "MM"
		}]`,
	"/total/country/belgium": `[
		{
			"Country": "Belgium",
			"CountryCode": "BE",
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
			"CountryCode": "BE",
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
		}]`,
	"/total/country/myanmar": `[
		{
			"Country": "Myanmar",
			"CountryCode": "",
			"Province": "",
			"City": "",
			"CityCode": "",
			"Lat": "0",
			"Lon": "0",
			"Confirmed": 8,
			"Deaths": 0,
			"Recovered": 0,
			"Active": 8,
 			"Date": "2020-01-31T00:00:00Z"
		}]`}

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

// makeClient returns a stubbed httpClient
func makeHTTPClient() *http.Client {
	rand.Seed(time.Now().UnixNano())
	return NewTestClient(func(req *http.Request) *http.Response {
		if rand.Intn(10) < 2 {
			return &http.Response{
				StatusCode: 429,
			}
		}
		response, ok := goodResponse[req.URL.Path]
		if ok == true {
			return &http.Response{
				StatusCode: 200,
				Header:     make(http.Header),
				Body:       ioutil.NopCloser(bytes.NewBufferString(response)),
			}
		}
		return &http.Response{
			StatusCode: 404,
		}
	})
}
