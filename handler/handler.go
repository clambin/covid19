package handler

import (
	"context"
	"fmt"
	"github.com/clambin/covid19/cache"
	populationStore "github.com/clambin/covid19/population/store"
	"github.com/clambin/grafana-json"
)

// Handler implements business logic for APIServer
type Handler struct {
	Cache           *cache.Cache
	PopulationStore populationStore.PopulationStore
}

// Targets for the Grafana SimpleJSON API Handler
var Targets = []string{
	"incremental",
	"cumulative",
	"evolution",
	"country-confirmed",
	"country-deaths",
	"country-confirmed-population",
	"country-deaths-population",
}

// required targets
//
// + confirmed / dead for a country per day
//   -> incremental
// + confirmed / dead for a country, cumulative
//   -> cumulative
// + confirmed / dead in total, per day
//   -> incremental
// + confirmed / dead in total, cumulative
//   -> cumulative
// - total confirmed by country
//   -> confirmed
// - total dead by country
//   -> dead
// - delta active by country (7-day average)
//   -> active
// - confirmed / dead per capita by country, cumulative
//   -> confirmed_capita
//   -> dead_capita

// Endpoints tells the server which endpoints we have implemented
func (handler *Handler) Endpoints() grafana_json.Endpoints {
	return grafana_json.Endpoints{
		Search:     handler.Search,
		TableQuery: handler.TableQuery,
		TagKeys:    handler.TagKeys,
		TagValues:  handler.TagValues,
	}
}

// Search returns all supported targets
func (handler *Handler) Search() []string {
	return Targets
}

// TagKeys returns all supported tag keys
func (handler *Handler) TagKeys(_ context.Context) []string {
	return []string{"Country Name"}
}

// TagValues returns all values for the specified tag
func (handler *Handler) TagValues(_ context.Context, key string) (values []string, err error) {
	if key != "Country Name" {
		return values, fmt.Errorf("unsupported tag '%s'", key)
	}

	return handler.Cache.DB.GetAllCountryNames()
}
