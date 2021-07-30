package backfill_test

import (
	"github.com/clambin/covid19/backfill"
	"github.com/clambin/covid19/coviddb"
	"github.com/clambin/covid19/coviddb/mock"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestBackFiller(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	covidDB := mock.Create([]coviddb.CountryEntry{
		{
			Timestamp: time.Date(2020, time.February, 1, 0, 0, 0, 0, time.UTC),
			Code:      "BE",
			Name:      "Belgium",
			Confirmed: 0,
			Deaths:    0,
			Recovered: 0,
		}})

	server := httptest.NewServer(http.HandlerFunc(covidAPI))
	defer server.Close()

	backFiller := backfill.Create(covidDB)
	backFiller.URL = server.URL

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

// covidAPI emulates the Covid API Server

func covidAPI(w http.ResponseWriter, req *http.Request) {
	// rand.Seed(time.Now().UnixNano())
	if rand.Intn(10) < 2 {
		http.Error(w, "slow down!", http.StatusTooManyRequests)
		return
	}

	response, ok := goodResponse[req.URL.Path]

	if ok == false {
		http.Error(w, "endpoint not implemented", http.StatusNotFound)
		return
	}

	_, _ = w.Write([]byte(response))
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
