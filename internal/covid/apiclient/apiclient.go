package apiclient

import (
	"encoding/json"
	"net/http"
	"time"
	// log "github.com/sirupsen/logrus"

	"github.com/clambin/gotools/rapidapi"
)

// API interface representing a Population API Client
type API interface {
	GetCountryStats() (map[string]CountryStats, error)
}

// APIClient API Client handle
type APIClient struct {
	rapidapi.Client
}

// New creates a new Covid API Client
func New(apiKey string) API {
	return &APIClient{rapidapi.Client{Client: &http.Client{}, HostName: rapidAPIHost, APIKey: apiKey}}
}

// CountryStats contains total figures for one country
type CountryStats struct {
	LastUpdate time.Time
	Confirmed  int64
	Deaths     int64
	Recovered  int64
}

// GetCountryStats finds the most recent figured for all countries
func (client *APIClient) GetCountryStats() (map[string]CountryStats, error) {
	countryStats := make(map[string]CountryStats, 0)
	stats, err := client.getStats()

	if err == nil {
		for _, entry := range stats.Data.Covid19Stats {
			mapEntry, ok := countryStats[entry.Country]
			if ok == false {
				mapEntry = CountryStats{Confirmed: 0, Deaths: 0, Recovered: 0}
				countryStats[entry.Country] = mapEntry
			}
			mapEntry.LastUpdate = entry.LastUpdate
			mapEntry.Confirmed += entry.Confirmed
			mapEntry.Deaths += entry.Deaths
			mapEntry.Recovered += entry.Recovered

			countryStats[entry.Country] = mapEntry
		}
	}

	return countryStats, err
}

//
// internal functions
//

const (
	rapidAPIHost = "covid-19-coronavirus-statistics.p.rapidapi.com"
)

// statsResponse matches the layout of the API's response object
// so json.Decoder will parse it directly into the struct
// !!! fields needs to start w/ uppercase or decoder will ignore them
type statsResponse struct {
	Error      bool
	StatusCode int
	Message    string
	Data       struct {
		LastChecked  time.Time
		Covid19Stats []struct {
			Country    string
			LastUpdate time.Time
			Confirmed  int64
			Deaths     int64
			Recovered  int64
		}
	}
}

// getStats retrieves today's covid19 country stats from rapidapi.com
func (client *APIClient) getStats() (*statsResponse, error) {
	var stats statsResponse

	resp, err := client.Client.CallAsReader("/v1/stats")
	if err == nil {
		decoder := json.NewDecoder(resp)
		err = decoder.Decode(&stats)
	}

	return &stats, err
}
