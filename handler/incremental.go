package handler

import (
	"github.com/clambin/covid19/cache"
	"github.com/clambin/covid19/models"
	grafana_json "github.com/clambin/grafana-json"
)

func (handler *Handler) handleIncremental(args *grafana_json.TableQueryArgs) (response *grafana_json.TableQueryResponse, err error) {
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

func (handler *Handler) getDeltasForCountry(args *grafana_json.TableQueryArgs) (deltas []cache.Entry, err error) {
	var countryName string
	countryName, err = evaluateAhHocFilter(args.AdHocFilters)

	if err != nil {
		return
	}

	var entries []*models.CountryEntry
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
