package probe

import (
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"net/http"
	"net/url"
)

// APIClient interface representing a Population API Client
type APIClient interface {
	GetPopulation(codes string) (int64, error)
}

// RapidAPIClient API Client handle
type RapidAPIClient struct {
	HTTPClient *http.Client
	APIKey     string
}

// NewAPIClient creates a new Population API Client
func NewAPIClient(apiKey string) APIClient {
	return &RapidAPIClient{HTTPClient: &http.Client{}, APIKey: apiKey}
}

// GetPopulation finds the most recent data for a countries
func (client *RapidAPIClient) GetPopulation(country string) (population int64, err error) {
	var stats populationResponse
	stats, err = client.getStats(country)

	if err == nil && stats.OK == false {
		err = fmt.Errorf("invalid response received from %s", rapidAPIHost)
	}

	if err == nil {
		population = stats.Body.Population
	}

	return
}

func (client *RapidAPIClient) GetCountries() (countries []string, err error) {
	myURL := "https://" + rapidAPIHost + "/allcountriesname"

	var req *http.Request
	req, _ = http.NewRequest(http.MethodGet, myURL, nil)
	req.Header.Add("x-rapidapi-key", client.APIKey)
	req.Header.Add("x-rapidapi-host", rapidAPIHost)

	var resp *http.Response
	resp, err = client.HTTPClient.Do(req)
	if err == nil && resp.StatusCode == http.StatusOK {
		var stats struct {
			OK   bool
			Body struct {
				Countries []string
			}
		}
		decoder := json.NewDecoder(resp.Body)
		err = decoder.Decode(&stats)
		_ = resp.Body.Close()

		if err == nil && stats.OK == true {
			countries = stats.Body.Countries
		}
	}

	return

}

//
// internal functions
//

const (
	rapidAPIHost = "world-population.p.rapidapi.com"
)

type populationResponse struct {
	OK   bool                   `json:"ok"`
	Body populationResponseBody `json:"body"`
}

type populationResponseBody struct {
	CountryName string  `json:"country_name"`
	Population  int64   `json:"population"`
	Ranking     int     `json:"ranking"`
	WorldShare  float32 `json:"world_share"`
}

// getStats retrieves today's covid19 country stats from rapidapi.com
func (client *RapidAPIClient) getStats(country string) (stats populationResponse, err error) {
	myURL := "https://" + rapidAPIHost + "/population?country_name=" + url.QueryEscape(country)

	var req *http.Request
	req, _ = http.NewRequest(http.MethodGet, myURL, nil)
	req.Header.Add("x-rapidapi-key", client.APIKey)
	req.Header.Add("x-rapidapi-host", rapidAPIHost)

	var resp *http.Response
	resp, err = client.HTTPClient.Do(req)

	log.WithError(err).Debugf("called %s", myURL)

	if err == nil {
		if resp.StatusCode == http.StatusOK {
			decoder := json.NewDecoder(resp.Body)
			err = decoder.Decode(&stats)
		} else {
			err = fmt.Errorf("%s", resp.Status)
		}
		_ = resp.Body.Close()
	}

	return
}
