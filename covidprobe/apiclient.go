package covidprobe

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/clambin/gotools/rapidapi"
	"time"
)

// APIClient interface representing a Population API Client
//go:generate mockery --name APIClient
type APIClient interface {
	GetCountryStats(ctx context.Context) (map[string]CountryStats, error)
}

// RapidAPIClient Client handle
type RapidAPIClient struct {
	rapidapi.API
}

// NewAPIClient creates a new Covid API Client
func NewAPIClient(apiKey string) *RapidAPIClient {
	return &RapidAPIClient{
		&rapidapi.Client{
			Hostname: rapidAPIHost,
			APIKey:   apiKey,
		},
	}
}

// CountryStats contains total figures for one country
type CountryStats struct {
	LastUpdate time.Time
	Confirmed  int64
	Deaths     int64
	Recovered  int64
}

// GetCountryStats finds the most recent figures for all countries
func (client *RapidAPIClient) GetCountryStats(ctx context.Context) (countryStats map[string]CountryStats, err error) {
	countryStats = make(map[string]CountryStats, 0)
	var stats statsResponse
	stats, err = client.getStats(ctx)

	if err != nil {
		return
	}

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
	return
}

//
// internal functions
//

const (
	rapidAPIHost = "covid-19-coronavirus-statistics.p.rapidapi.com"
)

// statsResponse matches the layout of the API's response object
// so json.Decoder will parse it directly into the struct
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
func (client *RapidAPIClient) getStats(ctx context.Context) (stats statsResponse, err error) {
	var body []byte
	body, err = client.API.CallWithContext(ctx, "/v1/stats")
	if err != nil {
		return
	}
	decoder := json.NewDecoder(bytes.NewReader(body))
	err = decoder.Decode(&stats)
	return
}
