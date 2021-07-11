package probe_test

import (
	"bytes"
	"fmt"
	"github.com/clambin/covid19/population/probe"
	"github.com/clambin/gotools/httpstub"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"testing"
)

func TestGetPopulation(t *testing.T) {
	apiClient := probe.NewAPIClient("1234")
	apiClient.(*probe.RapidAPIClient).Client.Client = httpstub.NewTestClient(serverStub)

	population, err := apiClient.GetPopulation("Belgium")

	assert.NoError(t, err)
	assert.Equal(t, int64(20), population)

	population, err = apiClient.GetPopulation("United States")

	assert.NoError(t, err)
	assert.Equal(t, int64(40), population)

	population, err = apiClient.GetPopulation("??")

	assert.Error(t, err)
}

func TestGetCountries(t *testing.T) {
	apiClient := probe.NewAPIClient("1234")
	apiClient.(*probe.RapidAPIClient).Client.Client = httpstub.NewTestClient(serverStub)

	countries, err := apiClient.GetCountries()
	assert.NoError(t, err)
	assert.Len(t, countries, 2)
	assert.Contains(t, countries, "Belgium")
	assert.Contains(t, countries, "United States")
}

func serverStub(req *http.Request) *http.Response {
	var response string
	if req.URL.Path == "/population" {
		switch req.URL.RawQuery {
		case "country_name=Belgium":
			response = fmt.Sprintf(countryResponse, "Belgium", 20)
		case "country_name=United+States":
			response = fmt.Sprintf(countryResponse, "United States", 40)
		case "country_name=Faeroe+Islands":
			response = fmt.Sprintf(countryResponse, "Faeroe Islands", 5)
		}
	} else if req.URL.Path == "/allcountriesname" {
		response = allCountriesResponse
	}

	if response == "" {
		return &http.Response{StatusCode: http.StatusNotFound}
	}

	return &http.Response{
		StatusCode: 200,
		Header:     make(http.Header),
		Body:       ioutil.NopCloser(bytes.NewBufferString(response)),
	}
}

var countryResponse = `
	{
		"ok": true,
		"body": {
			"country_name": "%s",
			"population": %d
		}
	}
`

var allCountriesResponse = `
	{
		"ok": true,
		"body": {
			"countries": [
				"Belgium",
				"United States"
			]
		}
	}
`
