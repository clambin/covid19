package covidprobe

import (
	"errors"
	"time"
	"net/http"
	"encoding/json"

	// log "github.com/sirupsen/logrus"
)

// CovidAPIClient API Client handle
type CovidAPIClient struct {
	client *http.Client
	apiKey  string
}

// Covid19CountryStats contains total figures for one country
type Covid19CountryStats struct {
	LastUpdate time.Time
	Confirmed  int64
	Deaths     int64
	Recovered int64
}

// NewCovidAPIClient creates a new Covid API Client
func NewCovidAPIClient(client *http.Client, apiKey string) (*CovidAPIClient) {
	return &CovidAPIClient{client: client, apiKey: apiKey}
}

// GetCountryStats finds the most recent figured for all countries
func (client *CovidAPIClient) GetCountryStats() (map[string]Covid19CountryStats, error) {
	countryStats := make(map[string]Covid19CountryStats, 0)
	stats, err := client.getStats()

	if err == nil {
		for _, entry := range stats.Data.Covid19Stats {
			mapEntry, ok := countryStats[entry.Country]
            if ok == false {
				mapEntry := Covid19CountryStats{Confirmed: 0, Deaths: 0,  Recovered: 0}
                countryStats[entry.Country] = mapEntry
            }
            mapEntry.LastUpdate  = entry.LastUpdate
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
	rapidAPIHost = string("covid-19-coronavirus-statistics.p.rapidapi.com")
	url =  string("https://") + rapidAPIHost
)

type covid19StatsResponse struct {
	Error                bool
	StatusCode           int
	Message              string
	Data struct {
		LastChecked      time.Time
		Covid19Stats   []struct{
			// City         string
			// Province     string
			Country      string
			LastUpdate   time.Time
			Confirmed    int64
			Deaths       int64
			Recovered    int64
		}
	}
}

// getStats retrieves today's covid19 country stats from rapidapi.com
func (client *CovidAPIClient) getStats() (*covid19StatsResponse, error) {
	req, _ := http.NewRequest("GET", url + "/v1/stats", nil)
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

	var stats covid19StatsResponse
	decoder := json.NewDecoder(resp.Body)
    err = decoder.Decode(&stats)
	return &stats, err
}

