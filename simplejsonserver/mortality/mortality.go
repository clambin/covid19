package mortality

import (
	"context"
	"fmt"
	"github.com/clambin/covid19/models"
	"github.com/clambin/simplejson/v6"
	"sort"
	"time"
)

// Handler returns the mortality by nr. of confirmed cases
type Handler struct {
	CovidDB CovidGetter
}

type CovidGetter interface {
	GetLatestForCountries(time2 time.Time) (map[string]models.CountryEntry, error)
}

var _ simplejson.Handler = &Handler{}

func (handler *Handler) Endpoints() (endpoints simplejson.Endpoints) {
	return simplejson.Endpoints{
		Query: handler.tableQuery,
	}
}

func (handler *Handler) tableQuery(_ context.Context, req simplejson.QueryRequest) (simplejson.Response, error) {
	entries, err := handler.CovidDB.GetLatestForCountries(req.Args.Range.To)
	if err != nil {
		return nil, fmt.Errorf("database: %w", err)
	}

	var countryNames []string
	for countryName := range entries {
		countryNames = append(countryNames, countryName)
	}
	sort.Strings(countryNames)

	var countryCodes []string
	var timestamps []time.Time
	var ratios []float64

	for _, countryName := range countryNames {
		entry := entries[countryName]
		timestamps = append(timestamps, entry.Timestamp)
		countryCodes = append(countryCodes, entry.Code)
		var ratio float64
		if entry.Confirmed > 0 {
			ratio = float64(entry.Deaths) / float64(entry.Confirmed)
		}
		ratios = append(ratios, ratio)
	}

	response := simplejson.TableResponse{Columns: []simplejson.Column{
		{Text: "timestamp", Data: simplejson.TimeColumn(timestamps)},
		{Text: "country", Data: simplejson.StringColumn(countryCodes)},
		{Text: "ratio", Data: simplejson.NumberColumn(ratios)},
	}}
	return &response, nil
}
