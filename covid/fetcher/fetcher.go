package fetcher

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/clambin/covid19/models"
	"github.com/clambin/go-rapidapi"
	log "github.com/sirupsen/logrus"
	"time"
)

// Fetcher retrieves COVID-19 stats from the rapidAPI server
//
//go:generate mockery --name Fetcher
type Fetcher interface {
	GetCountryStats(ctx context.Context) (countryEntry []models.CountryEntry, err error)
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
	invalidCountries StringSet
}

// GetCountryStats called the API to retrieve the latest COVID-19 stats
func (client *Client) GetCountryStats(ctx context.Context) (countryEntry []models.CountryEntry, err error) {
	var stats statsResponse
	stats, err = client.getStats(ctx)

	if err != nil {
		return
	}

	countryEntry = client.filterUnsupportedCountries(&stats)
	countryEntry = sumByCountry(countryEntry)
	return
}

func (client *Client) filterUnsupportedCountries(stats *statsResponse) (entries []models.CountryEntry) {
	for _, entry := range stats.Data.Covid19Stats {
		code, found := CountryCodes[entry.Country]

		if !found {
			if found = client.invalidCountries.Set(entry.Country); !found {
				log.WithField("name", entry.Country).Warning("unknown country name received from COVID-19 API")
				continue
			}
		}
		entries = append(entries, models.CountryEntry{
			Timestamp: entry.LastUpdate,
			Code:      code,
			Name:      entry.Country,
			Confirmed: entry.Confirmed,
			Recovered: entry.Recovered,
			Deaths:    entry.Deaths,
		})
	}
	return
}

func sumByCountry(entries []models.CountryEntry) (sum []models.CountryEntry) {
	summed := make(map[string]models.CountryEntry)

	for _, entry := range entries {
		sumEntry, found := summed[entry.Code]
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
		summed[entry.Code] = sumEntry
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

func (client *Client) getStats(ctx context.Context) (stats statsResponse, err error) {
	const endpoint = "/v1/stats"

	var body []byte
	body, err = client.API.CallWithContext(ctx, endpoint)

	if err == nil {
		err = json.NewDecoder(bytes.NewReader(body)).Decode(&stats)
	}

	return
}
