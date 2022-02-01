package countries

import (
	"context"
	covidStore "github.com/clambin/covid19/covid/store"
	"github.com/clambin/covid19/models"
	"github.com/clambin/simplejson/v2"
	"github.com/clambin/simplejson/v2/query"
	"time"
)

const (
	CountryConfirmed = iota
	CountryDeaths
)

// ByCountryHandler returns the latest stats by country
type ByCountryHandler struct {
	CovidDB covidStore.CovidStore
	Mode    int
}

var _ simplejson.Handler = &ByCountryHandler{}

func (handler ByCountryHandler) Endpoints() (endpoints simplejson.Endpoints) {
	return simplejson.Endpoints{
		TableQuery: handler.tableQuery,
	}
}

func (handler *ByCountryHandler) tableQuery(_ context.Context, args query.Args) (response *query.TableResponse, err error) {
	return getStatsByCountry(handler.CovidDB, args, handler.Mode)
}

func getStatsByCountry(db covidStore.CovidStore, args query.Args, mode int) (response *query.TableResponse, err error) {
	var names []string
	names, err = db.GetAllCountryNames()

	if err != nil {
		return
	}

	var entries map[string]models.CountryEntry
	if args.Range.To.IsZero() {
		entries, err = db.GetLatestForCountries(names)
	} else {
		entries, err = db.GetLatestForCountriesByTime(names, args.Range.To)
	}

	if err != nil {
		return
	}

	var timestamp time.Time
	response = &query.TableResponse{}

	for _, name := range names {
		entry := entries[name]
		if timestamp.IsZero() {
			timestamp = entry.Timestamp
			response.Columns = append(response.Columns, query.Column{
				Text: "timestamp",
				Data: query.TimeColumn([]time.Time{timestamp}),
			})
		}

		var value float64
		switch mode {
		case CountryConfirmed:
			value = float64(entry.Confirmed)
		case CountryDeaths:
			value = float64(entry.Deaths)
		}

		response.Columns = append(response.Columns, query.Column{
			Text: name,
			Data: query.NumberColumn([]float64{value}),
		})
	}

	return
}