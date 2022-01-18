package handler

import (
	"github.com/clambin/covid19/cache"
	"github.com/clambin/covid19/models"
	"github.com/clambin/simplejson"
)

func (handler *Handler) handleCumulative(args *simplejson.TableQueryArgs) (response *simplejson.TableQueryResponse, err error) {
	var totals []cache.Entry
	if len(args.AdHocFilters) > 0 {
		totals, err = handler.getTotalsForCountry(args)
	} else {
		totals, err = handler.Cache.GetTotals(args.Range.To)
	}

	if err == nil {
		response = buildResponse(totals, args.Range)
	}

	return
}

func (handler *Handler) getTotalsForCountry(args *simplejson.TableQueryArgs) (totals []cache.Entry, err error) {
	var countryName string
	countryName, err = evaluateAhHocFilter(args.AdHocFilters)

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
