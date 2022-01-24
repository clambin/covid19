package summarized

import (
	"context"
	"fmt"
	"github.com/clambin/covid19/cache"
	"github.com/clambin/covid19/models"
	"github.com/clambin/simplejson/v2"
	"github.com/clambin/simplejson/v2/query"
)

// IncrementalHandler returns the incremental number of cases & deaths. If an adhoc filter exists, it returns the
// incremental cases/deaths for that country
type IncrementalHandler struct {
	Cache *cache.Cache
}

var _ simplejson.Handler = &IncrementalHandler{}

func (handler IncrementalHandler) Endpoints() (endpoints simplejson.Endpoints) {
	return simplejson.Endpoints{
		TableQuery: handler.tableQuery,
		TagKeys:    handler.tagKeys,
		TagValues:  handler.tagValues,
	}
}

func (handler *IncrementalHandler) tableQuery(_ context.Context, args query.Args) (response *query.TableResponse, err error) {
	var deltas []cache.Entry
	if len(args.AdHocFilters) > 0 {
		deltas, err = handler.getDeltasForCountry(args)
	} else {
		deltas, err = handler.Cache.GetDeltas(args.Range.To)
	}

	if err == nil {
		response = buildResponse(deltas, args.Range)
	}
	return
}

func (handler *IncrementalHandler) getDeltasForCountry(args query.Args) (deltas []cache.Entry, err error) {
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

	var confirmed, deaths int64
	for _, entry := range entries {
		deltas = append(deltas, cache.Entry{
			Timestamp: entry.Timestamp,
			Confirmed: entry.Confirmed - confirmed,
			Deaths:    entry.Deaths - deaths,
		})
		confirmed = entry.Confirmed
		deaths = entry.Deaths
	}
	return
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
