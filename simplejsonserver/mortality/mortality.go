package mortality

import (
	"context"
	covidStore "github.com/clambin/covid19/db"
	"github.com/clambin/covid19/models"
	"github.com/clambin/simplejson/v5"
	"time"
)

// Handler returns the mortality by nr. of confirmed cases
type Handler struct {
	CovidDB covidStore.CovidStore
}

var _ simplejson.Handler = &Handler{}

func (handler *Handler) Endpoints() (endpoints simplejson.Endpoints) {
	return simplejson.Endpoints{
		Query: handler.tableQuery,
	}
}

func (handler *Handler) tableQuery(_ context.Context, req simplejson.QueryRequest) (response simplejson.Response, err error) {
	var countryNames []string
	countryNames, err = handler.CovidDB.GetAllCountryNames()
	if err != nil {
		return
	}

	var entries map[string]models.CountryEntry
	entries, err = handler.CovidDB.GetLatestForCountriesByTime(countryNames, req.Args.Range.To)
	if err != nil {
		return
	}

	var timestamps []time.Time
	var countryCodes []string
	var ratios []float64

	for _, countryName := range countryNames {
		entry, found := entries[countryName]
		if !found {
			continue
		}

		timestamps = append(timestamps, entry.Timestamp)
		countryCodes = append(countryCodes, entry.Code)
		var ratio float64
		if entry.Confirmed > 0 {
			ratio = float64(entry.Deaths) / float64(entry.Confirmed)
		}
		ratios = append(ratios, ratio)
	}

	return &simplejson.TableResponse{Columns: []simplejson.Column{
		{Text: "timestamp", Data: simplejson.TimeColumn(timestamps)},
		{Text: "country", Data: simplejson.StringColumn(countryCodes)},
		{Text: "ratio", Data: simplejson.NumberColumn(ratios)},
	}}, nil
}
