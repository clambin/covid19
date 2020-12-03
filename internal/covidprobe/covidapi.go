package covidprobe

import (
	"time"
	"net/http"
	"encoding/json"
)

type CovidAPIClientIFC interface {
	GetLatestStats()
}

type CovidAPIClient struct {
	client *http.Client
	apiKey  string
}

const (
	url = string("https://covid-19-coronavirus-statistics.p.rapidapi.com")
)

// NewCovidAPIClient creates a new Covid API Client
func NewCovidAPIClient(client *http.Client, apiKey string) (*CovidAPIClient) {
	return &CovidAPIClient{client: client, apiKey: apiKey}
}

type Covid19StatsResponse struct {
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
			Confirmed    int
			Deaths       int
			Recovered    int
		}
	}
}

// Call the API
func (client *CovidAPIClient) GetStats() (*Covid19StatsResponse, error){
	req, _ := http.NewRequest("GET", url + "/v1/stats", nil)
	req.Header.Add("x-rapidapi-key", client.apiKey)
	req.Header.Add("x-rapidapi-host", "covid-19-coronavirus-statistics.p.rapidapi.com")

	resp, err := client.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var stats Covid19StatsResponse
	decoder := json.NewDecoder(resp.Body)
    err = decoder.Decode(&stats)
	return &stats, err
}

