package countries

import (
	"context"
	covidStore "github.com/clambin/covid19/covid/store"
	"github.com/clambin/covid19/models"
	"github.com/clambin/simplejson"
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

func (handler *ByCountryHandler) tableQuery(_ context.Context, args *simplejson.TableQueryArgs) (response *simplejson.TableQueryResponse, err error) {
	return getStatsByCountry(handler.CovidDB, args, handler.Mode)
}

func getStatsByCountry(db covidStore.CovidStore, args *simplejson.TableQueryArgs, mode int) (response *simplejson.TableQueryResponse, err error) {
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
	response = &simplejson.TableQueryResponse{}

	for _, name := range names {
		entry := entries[name]
		if timestamp.IsZero() {
			timestamp = entry.Timestamp
			response.Columns = append(response.Columns, simplejson.TableQueryResponseColumn{
				Text: "timestamp",
				Data: simplejson.TableQueryResponseTimeColumn([]time.Time{timestamp}),
			})
		}

		var value float64
		switch mode {
		case CountryConfirmed:
			value = float64(entry.Confirmed)
		case CountryDeaths:
			value = float64(entry.Deaths)
		}

		response.Columns = append(response.Columns, simplejson.TableQueryResponseColumn{
			Text: name,
			Data: simplejson.TableQueryResponseNumberColumn([]float64{value}),
		})
	}

	return
}
