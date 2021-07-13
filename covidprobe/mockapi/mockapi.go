package mockapi

import (
	"context"
	"github.com/clambin/covid19/covidprobe"
)

// API handle
type API struct {
	data map[string]covidprobe.CountryStats
}

// New creates a mockapi population API Client
func New(data map[string]covidprobe.CountryStats) *API {
	return &API{data: data}
}

// GetCountryStats returns the provided data
func (api *API) GetCountryStats(_ context.Context) (map[string]covidprobe.CountryStats, error) {
	return api.data, nil
}
