package handler

import (
	"github.com/clambin/covid19/cache"
	"github.com/clambin/covid19/models"
	grafana_json "github.com/clambin/grafana-json"
)

func (handler *Handler) handleCumulative(args *grafana_json.TableQueryArgs) (response *grafana_json.TableQueryResponse, err error) {
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

func (handler *Handler) getTotalsForCountry(args *grafana_json.TableQueryArgs) (totals []cache.Entry, err error) {
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
