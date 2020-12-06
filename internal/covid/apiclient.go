package covid

import (
	"errors"
	"time"
	"net/http"
	"encoding/json"

	// log "github.com/sirupsen/logrus"
)

// APIClient API Client handle
type APIClient struct {
	client *http.Client
	apiKey  string
}

// NewAPIClient creates a new Covid API Client
func NewAPIClient(client *http.Client, apiKey string) (*APIClient) {
	return &APIClient{client: client, apiKey: apiKey}
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
				mapEntry = CountryStats{Confirmed: 0, Deaths: 0,  Recovered: 0}
                countryStats[entry.Country] = mapEntry
            }
            mapEntry.LastUpdate  = entry.LastUpdate
            mapEntry.Confirmed  += entry.Confirmed
            mapEntry.Deaths     += entry.Deaths
            mapEntry.Recovered  += entry.Recovered

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

// statsResponse matches the layout of the API's response object
// so json.Decoder will parse it directly into the struct
// !!! fields needs to start w/ uppercase or decoder will ignore them
type statsResponse struct {
	Error                bool
	StatusCode           int
	Message              string
	Data struct {
		LastChecked      time.Time
		Covid19Stats   []struct{
			Country      string
			LastUpdate   time.Time
			Confirmed    int64
			Deaths       int64
			Recovered    int64
		}
	}
}

// getStats retrieves today's covid19 country stats from rapidapi.com
func (client *APIClient) getStats() (*statsResponse, error) {
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

	var stats statsResponse
	decoder := json.NewDecoder(resp.Body)
    err = decoder.Decode(&stats)
	return &stats, err
}

