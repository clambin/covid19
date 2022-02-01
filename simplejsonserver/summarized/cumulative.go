package summarized

import (
	"context"
	"fmt"
	"github.com/clambin/covid19/cache"
	"github.com/clambin/covid19/models"
	"github.com/clambin/simplejson/v3"
	"github.com/clambin/simplejson/v3/query"
)

// CumulativeHandler returns the incremental number of cases & deaths. If an adhoc filter exists, it returns the
// cumulative cases/deaths for that country
type CumulativeHandler struct {
	Cache *cache.Cache
}

var _ simplejson.Handler = &CumulativeHandler{}

func (handler CumulativeHandler) Endpoints() (endpoints simplejson.Endpoints) {
	return simplejson.Endpoints{
		Query:     handler.tableQuery,
		TagKeys:   handler.tagKeys,
		TagValues: handler.tagValues,
	}
}

func (handler *CumulativeHandler) tableQuery(_ context.Context, req query.Request) (response query.Response, err error) {
	var totals []cache.Entry
	if len(req.Args.AdHocFilters) > 0 {
		totals, err = handler.getTotalsForCountry(req.Args)
	} else {
		totals, err = handler.Cache.GetTotals(req.Args.Range.To)
	}

	if err == nil {
		response = buildResponse(totals, req.Args.Range)
	}
	return
}

func (handler *CumulativeHandler) getTotalsForCountry(args query.Args) (totals []cache.Entry, err error) {
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

	for _, entry := range entries {
		totals = append(totals, cache.Entry{
			Timestamp: entry.Timestamp,
			Confirmed: entry.Confirmed,
			Deaths:    entry.Deaths,
		})
	}
	return
}

func (handler *CumulativeHandler) tagKeys(_ context.Context) []string {
	return []string{"Country Name"}
}

func (handler *CumulativeHandler) tagValues(_ context.Context, key string) (values []string, err error) {
	if key != "Country Name" {
		return values, fmt.Errorf("unsupported tag '%s'", key)
	}

	return handler.Cache.DB.GetAllCountryNames()
}
