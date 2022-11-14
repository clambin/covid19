package backfill_test

import (
	"github.com/clambin/covid19/backfill"
	"github.com/clambin/covid19/db/mocks"
	"github.com/clambin/covid19/models"
	"github.com/stretchr/testify/require"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestBackfiller_Run(t *testing.T) {
	db := mocks.NewCovidStore(t)

	server := httptest.NewServer(http.HandlerFunc(covidAPI))
	defer server.Close()

	backFiller := backfill.New(db)
	backFiller.URL = server.URL

	db.On("Add", []models.CountryEntry{
		{
			Timestamp: time.Date(2020, 1, 23, 0, 0, 0, 0, time.UTC),
			Name:      "Belgium",
			Code:      "BE",
			Confirmed: 0,
			Recovered: 0,
			Deaths:    0,
		},
		{
			Timestamp: time.Date(2020, 2, 5, 0, 0, 0, 0, time.UTC),
			Name:      "Belgium",
			Code:      "BE",
			Confirmed: 1,
			Recovered: 0,
			Deaths:    0,
		},
	}).Return(nil)
	db.On("Add", []models.CountryEntry{{
		Timestamp: time.Date(2020, 2, 1, 0, 0, 0, 0, time.UTC),
		Name:      "Burma",
		Code:      "MM",
		Confirmed: 8,
		Recovered: 0,
		Deaths:    0,
	},
	}).Return(nil)

	err := backFiller.Run()
	require.NoError(t, err)
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
