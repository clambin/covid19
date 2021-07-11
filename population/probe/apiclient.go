package probe

import (
	"encoding/json"
	"fmt"
	"github.com/clambin/gotools/rapidapi"
	"io"
	"net/http"
	"net/url"
)

// APIClient interface representing a Population API Client
type APIClient interface {
	GetPopulation(codes string) (int64, error)
	GetCountries() ([]string, error)
}

// RapidAPIClient API Client handle
type RapidAPIClient struct {
	rapidapi.Client
}

const rapidAPIHost = "world-population.p.rapidapi.com"

// NewAPIClient creates a new Population API Client
func NewAPIClient(apiKey string) APIClient {
	return &RapidAPIClient{
		Client: rapidapi.Client{
			Client:   &http.Client{},
			HostName: rapidAPIHost,
			APIKey:   apiKey,
		},
	}
}

type populationResponse struct {
	OK   bool                   `json:"ok"`
	Body populationResponseBody `json:"body"`
}

type populationResponseBody struct {
	CountryName string  `json:"country_name"`
	Population  int64   `json:"population"`
	Ranking     int     `json:"ranking"`
	WorldShare  float32 `json:"world_share"`
}

// GetPopulation finds the most recent data for a countries
func (client *RapidAPIClient) GetPopulation(country string) (population int64, err error) {
	var resp io.Reader
	resp, err = client.Client.CallAsReader("/population?country_name=" + url.QueryEscape(country))

	var stats populationResponse
	if err == nil {
		decoder := json.NewDecoder(resp)
		err = decoder.Decode(&stats)
	}

	if err == nil {
		if stats.OK {
			population = stats.Body.Population
		} else {
			err = fmt.Errorf("invalid response received from %s", rapidAPIHost)
		}
	}

	return
}

func (client *RapidAPIClient) GetCountries() (countries []string, err error) {
	var resp io.Reader
	resp, err = client.Client.CallAsReader("/allcountriesname")

	var stats struct {
		OK   bool
		Body struct {
			Countries []string
		}
	}
	if err == nil {
		decoder := json.NewDecoder(resp)
		err = decoder.Decode(&stats)
	}

	if err == nil {
		if stats.OK == true {
			countries = stats.Body.Countries
		} else {
			err = fmt.Errorf("invalid response received from %s", rapidAPIHost)
		}
	}

	return
}
