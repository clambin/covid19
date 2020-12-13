package apiclient

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
)

// API interface representing a Population API Client
type API interface {
	GetPopulation() (map[string]int64, error)
}

// APIClient API Client handle
type APIClient struct {
	client *http.Client
	apiKey string
}

// New creates a new Population API Client
func New(apiKey string) API {
	return NewWithHTTPClient(&http.Client{}, apiKey)
}

// NewWithHTTPClient creates a new Covid API Client with a specified http.Client
// Used to stub the HTTP Server
func NewWithHTTPClient(client *http.Client, apiKey string) *APIClient {
	return &APIClient{client: client, apiKey: apiKey}
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
	rapidAPIHost = string("geohub3.p.rapidapi.com")
	url          = string("https://") + rapidAPIHost
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
	req, _ := http.NewRequest("GET", url+"/countries", nil)
	req.Header.Add("x-rapidapi-key", client.apiKey)
	req.Header.Add("x-rapidapi-host", rapidAPIHost)

	resp, err := client.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, errors.New(resp.Status)
	}

	var stats populationResponse
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&stats)

	return &stats, err
}
