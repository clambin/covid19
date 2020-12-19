package apiclient

import (
	"covid19/pkg/rapidapi"
	"encoding/json"
	"net/http"
	"strconv"
)

// API interface representing a Population API Client
type API interface {
	GetPopulation() (map[string]int64, error)
}

// APIClient API Client handle
type APIClient struct {
	rapidapi.Client
}

// New creates a new Population API Client
func New(apiKey string) API {
	return NewWithHTTPClient(&http.Client{}, apiKey)
}

// NewWithHTTPClient creates a new Covid API Client with a specified http.Client
// Used to stub the HTTP Server
func NewWithHTTPClient(client *http.Client, apiKey string) *APIClient {
	return &APIClient{rapidapi.Client{Client: client, HostName: rapidAPIHost, APIKey: apiKey}}
}

// GetPopulation finds the most recent figured for all countries
func (client *APIClient) GetPopulation() (map[string]int64, error) {
	entries := make(map[string]int64, 0)
	stats, err := client.getStats()

	if err == nil {
		for _, entry := range stats.Data.Countries {
			population, err := strconv.ParseInt(entry.Population, 10, 64)
			if err == nil {
				entries[entry.CountryCode] = population
			}
		}
	}

	return entries, err
}

//
// internal functions
//

const (
	rapidAPIHost = "geohub3.p.rapidapi.com"
)

type populationResponse struct {
	Data struct {
		Countries []struct {
			CountryCode string
			Population  string
		}
	}
}

// getStats retrieves today's covid19 country stats from rapidapi.com
func (client *APIClient) getStats() (*populationResponse, error) {
	var stats populationResponse

	resp, err := client.Client.CallAsReader("/countries")
	if err == nil {
		decoder := json.NewDecoder(resp)
		err = decoder.Decode(&stats)
	}

	return &stats, err
}
