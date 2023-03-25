package fetcher

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/clambin/covid19/models"
	"github.com/clambin/go-rapidapi"
	"time"
)

// Fetcher retrieves COVID-19 stats from the rapidAPI server
//
//go:generate mockery --name Fetcher
type Fetcher interface {
	Fetch(ctx context.Context) (countryEntry []models.CountryEntry, err error)
}

// CountryStats contains total figures for one country
type CountryStats struct {
	LastUpdate time.Time
	Confirmed  int64
	Deaths     int64
	Recovered  int64
}

var _ Fetcher = &Client{}

// Client implements the Fetcher interface
type Client struct {
	rapidapi.API
}

// Fetch called the API to retrieve the latest (raw) COVID-19 stats
func (client *Client) Fetch(ctx context.Context) ([]models.CountryEntry, error) {
	stats, err := client.getStats(ctx)
	if err != nil {
		return nil, err
	}
	var records []models.CountryEntry
	for _, entry := range stats.Data.Covid19Stats {
		records = append(records, models.CountryEntry{
			Timestamp: entry.LastUpdate.UTC(),
			Name:      entry.Country,
			Confirmed: entry.Confirmed,
			Recovered: entry.Recovered,
			Deaths:    entry.Deaths,
		})
	}
	return sumByCountry(records), nil
}

func sumByCountry(entries []models.CountryEntry) (sum []models.CountryEntry) {
	summed := make(map[string]models.CountryEntry)

	for _, entry := range entries {
		sumEntry, found := summed[entry.Name]
		if !found {
			sumEntry = models.CountryEntry{
				Timestamp: entry.Timestamp,
				Code:      entry.Code,
				Name:      entry.Name,
			}
		}
		sumEntry.Confirmed += entry.Confirmed
		sumEntry.Recovered += entry.Recovered
		sumEntry.Deaths += entry.Deaths
		summed[entry.Name] = sumEntry
	}

	for _, entry := range summed {
		sum = append(sum, entry)
	}

	return
}

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

func (client *Client) getStats(ctx context.Context) (statsResponse, error) {
	const endpoint = "/v1/stats"
	var stats statsResponse
	body, err := client.API.CallWithContext(ctx, endpoint)
	if err == nil {
		err = json.NewDecoder(bytes.NewReader(body)).Decode(&stats)
	}
	return stats, err
}
