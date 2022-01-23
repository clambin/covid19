package mortality

import (
	"context"
	covidStore "github.com/clambin/covid19/covid/store"
	"github.com/clambin/covid19/models"
	"github.com/clambin/simplejson"
	"time"
)

// Handler returns the mortality by nr. of confirmed cases
type Handler struct {
	CovidDB covidStore.CovidStore
}

var _ simplejson.Handler = &Handler{}

func (handler Handler) Endpoints() (endpoints simplejson.Endpoints) {
	return simplejson.Endpoints{
		TableQuery: handler.tableQuery,
	}
}

func (handler *Handler) tableQuery(_ context.Context, args *simplejson.TableQueryArgs) (response *simplejson.TableQueryResponse, err error) {
	var countryNames []string
	countryNames, err = handler.CovidDB.GetAllCountryNames()
	if err != nil {
		return
	}

	var entries map[string]models.CountryEntry
	entries, err = handler.CovidDB.GetLatestForCountriesByTime(countryNames, args.Range.To)
	if err != nil {
		return
	}

	var timestamps []time.Time
	var countryCodes []string
	var ratios []float64

	for _, countryName := range countryNames {
		entry, ok := entries[countryName]
		if ok == false {
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

	return &simplejson.TableQueryResponse{Columns: []simplejson.TableQueryResponseColumn{
		{
			Text: "timestamp",
			Data: simplejson.TableQueryResponseTimeColumn(timestamps),
		},
		{
			Text: "country",
			Data: simplejson.TableQueryResponseStringColumn(countryCodes),
		},
		{
			Text: "ratio",
			Data: simplejson.TableQueryResponseNumberColumn(ratios),
		},
	}}, nil
}
