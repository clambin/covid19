package backfill_test

import (
	"github.com/clambin/covid19/backfill"
	"github.com/clambin/covid19/internal/testtools/db/covid"
	"github.com/clambin/covid19/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestBackfiller_Run(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(covidAPI))
	defer server.Close()
	store := covid.FakeStore{}

	backFiller := backfill.New(&store)
	backFiller.Client = backfill.Client{URL: server.URL}

	err := backFiller.Run()
	require.NoError(t, err)

	content, _ := store.GetAllForRange(time.Time{}, time.Time{})
	assert.Equal(t, []models.CountryEntry{
		{Timestamp: time.Date(2020, time.January, 23, 0, 0, 0, 0, time.UTC), Code: "BE", Name: "Belgium", Confirmed: 0, Recovered: 0, Deaths: 0},
		{Timestamp: time.Date(2020, time.February, 1, 0, 0, 0, 0, time.UTC), Code: "MM", Name: "Burma", Confirmed: 8, Recovered: 0, Deaths: 0},
		{Timestamp: time.Date(2020, time.February, 5, 0, 0, 0, 0, time.UTC), Code: "BE", Name: "Belgium", Confirmed: 1, Recovered: 0, Deaths: 0},
	}, content)
}

// covidAPI emulates the Covid API Server
func covidAPI(w http.ResponseWriter, req *http.Request) {
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
		}]`,
}
