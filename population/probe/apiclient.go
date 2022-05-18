package probe

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/clambin/go-rapidapi"
	"net/url"
)

// APIClient interface representing a Population API Client
//go:generate mockery --name APIClient
type APIClient interface {
	GetPopulation(ctx context.Context, codes string) (int64, error)
	GetCountries(ctx context.Context) ([]string, error)
}

// RapidAPIClient API Client handle
type RapidAPIClient struct {
	rapidapi.API
}

const rapidAPIHost = "world-population.p.rapidapi.com"

// NewAPIClient creates a new Population API Client
func NewAPIClient(apiKey string) *RapidAPIClient {
	return &RapidAPIClient{
		API: rapidapi.New(rapidAPIHost, apiKey),
	}
}

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

// GetPopulation finds the most recent data for a country
func (client *RapidAPIClient) GetPopulation(ctx context.Context, country string) (population int64, err error) {
	var body []byte
	body, err = client.Call(ctx, "/population?country_name="+url.QueryEscape(country))

	var stats populationResponse
	if err == nil {
		decoder := json.NewDecoder(bytes.NewReader(body))
		err = decoder.Decode(&stats)
	}

	if err == nil {
		if stats.OK {
			population = stats.Body.Population
		} else {
			err = fmt.Errorf("invalid response received from %s", rapidAPIHost)
		}
	}

	return
}

// GetCountries returns all country names that the RapidAPI API supports
func (client *RapidAPIClient) GetCountries(ctx context.Context) (countries []string, err error) {
	var body []byte
	body, err = client.Call(ctx, "/allcountriesname")

	var stats struct {
		OK   bool
		Body struct {
			Countries []string
		}
	}

	if err == nil {
		err = json.NewDecoder(bytes.NewReader(body)).Decode(&stats)
	}

	if err == nil {
		if stats.OK {
			countries = stats.Body.Countries
		} else {
			err = fmt.Errorf("invalid response received from %s", rapidAPIHost)
		}
	}

	return
}

// Call calls the Population API for the provided endpoint
func (client *RapidAPIClient) Call(ctx context.Context, endpoint string) (body []byte, err error) {
	return client.API.CallWithContext(ctx, endpoint)
}
