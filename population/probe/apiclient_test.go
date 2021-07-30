package probe_test

import (
	"context"
	"fmt"
	"github.com/clambin/covid19/population/probe"
	"github.com/stretchr/testify/assert"
	"html"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetPopulation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(serverStub))
	defer server.Close()

	apiClient := probe.NewAPIClient("1234")
	apiClient.(*probe.RapidAPIClient).Client.URL = server.URL

	ctx := context.Background()
	population, err := apiClient.GetPopulation(ctx, "Belgium")

	assert.NoError(t, err)
	assert.Equal(t, int64(20), population)

	population, err = apiClient.GetPopulation(ctx, "United States")

	assert.NoError(t, err)
	assert.Equal(t, int64(40), population)

	population, err = apiClient.GetPopulation(ctx, "??")

	assert.Error(t, err)
}

func TestGetCountries(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(serverStub))
	defer server.Close()

	apiClient := probe.NewAPIClient("1234")
	apiClient.(*probe.RapidAPIClient).Client.URL = server.URL

	ctx := context.Background()
	countries, err := apiClient.GetCountries(ctx)
	assert.NoError(t, err)
	assert.Len(t, countries, 2)
	assert.Contains(t, countries, "Belgium")
	assert.Contains(t, countries, "United States")
}

func serverStub(w http.ResponseWriter, req *http.Request) {
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
		http.Error(w, "endpoint not implemented:"+html.EscapeString(req.URL.Path), http.StatusNotFound)
	}

	_, _ = w.Write([]byte(response))
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
