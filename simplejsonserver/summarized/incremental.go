package summarized

import (
	"context"
	"fmt"
	"github.com/clambin/covid19/cache"
	"github.com/clambin/covid19/models"
	"github.com/clambin/simplejson/v3"
	"github.com/clambin/simplejson/v3/data"
	"github.com/clambin/simplejson/v3/query"
	"time"
)

// IncrementalHandler returns the incremental number of cases & deaths. If an adhoc filter exists, it returns the
// incremental cases/deaths for that country
type IncrementalHandler struct {
	Cache *cache.Cache
}

var _ simplejson.Handler = &IncrementalHandler{}

func (handler IncrementalHandler) Endpoints() (endpoints simplejson.Endpoints) {
	return simplejson.Endpoints{
		Query:     handler.tableQuery,
		TagKeys:   handler.tagKeys,
		TagValues: handler.tagValues,
	}
}

func (handler *IncrementalHandler) tableQuery(_ context.Context, req query.Request) (response query.Response, err error) {
	var deltas *data.Table
	if len(req.Args.AdHocFilters) > 0 {
		deltas, err = handler.getDeltasForCountry(req.Args)
		if err == nil {
			deltas = deltas.Filter(req.Args)
		}
	} else {
		deltas, err = handler.Cache.GetDeltas(req.Args.Range.To)
	}

	if err == nil {
		response = deltas.CreateTableResponse()
	}
	return
}

func (handler *IncrementalHandler) getDeltasForCountry(args query.Args) (deltas *data.Table, err error) {
	var countryName string
	countryName, err = evaluateAdHocFilter(args.AdHocFilters)

	if err != nil {
		return
	}

	var entries []models.CountryEntry
	entries, err = handler.Cache.DB.GetAllForCountryName(countryName)

	if err != nil {
		return
	}

	timestamps := make([]time.Time, len(entries))
	confirmed := make([]float64, len(entries))
	deaths := make([]float64, len(entries))

	var lastConfirmed, lastDeaths float64
	for idx, entry := range entries {
		timestamps[idx] = entry.Timestamp
		confirmed[idx] = float64(entry.Confirmed) - lastConfirmed
		deaths[idx] = float64(entry.Deaths) - lastDeaths
		lastConfirmed = float64(entry.Confirmed)
		lastDeaths = float64(entry.Deaths)
	}

	return data.New(
		data.Column{Name: "timestamp", Values: timestamps},
		data.Column{Name: "confirmed", Values: confirmed},
		data.Column{Name: "deaths", Values: deaths},
	), nil
}

func (handler *IncrementalHandler) tagKeys(_ context.Context) []string {
	return []string{"Country Name"}
}

func (handler *IncrementalHandler) tagValues(_ context.Context, key string) (values []string, err error) {
	if key != "Country Name" {
		return values, fmt.Errorf("unsupported tag '%s'", key)
	}

	return handler.Cache.DB.GetAllCountryNames()
}
