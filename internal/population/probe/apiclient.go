package probe

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/clambin/gotools/rapidapi"
)

// APIClient interface representing a Population API Client
type APIClient interface {
	GetPopulation() (map[string]int64, error)
}

// RapidAPIClient API Client handle
type RapidAPIClient struct {
	rapidapi.Client
}

// New creates a new Population API Client
func NewAPIClient(apiKey string) APIClient {
	return &RapidAPIClient{rapidapi.Client{Client: &http.Client{}, HostName: rapidAPIHost, APIKey: apiKey}}
}

// GetPopulation finds the most recent data for all countries
func (client *RapidAPIClient) GetPopulation() (map[string]int64, error) {
	entries := make(map[string]int64)
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
func (client *RapidAPIClient) getStats() (*populationResponse, error) {
	var stats populationResponse

	resp, err := client.Client.CallAsReader("/countries")
	if err == nil {
		decoder := json.NewDecoder(resp)
		err = decoder.Decode(&stats)
	}

	return &stats, err
}
