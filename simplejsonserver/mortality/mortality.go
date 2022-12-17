package mortality

import (
	"context"
	"fmt"
	covidStore "github.com/clambin/covid19/db"
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

func (handler *Handler) tableQuery(_ context.Context, req simplejson.QueryRequest) (simplejson.Response, error) {
	entries, err := handler.CovidDB.GetLatestForCountriesByTime(req.Args.Range.To)
	if err != nil {
		return nil, fmt.Errorf("database: %w", err)
	}

	var timestamps []time.Time
	var countryCodes []string
	var ratios []float64

	for _, entry := range entries {
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
