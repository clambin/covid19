package countries

import (
	"context"
	covidStore "github.com/clambin/covid19/covid/store"
	"github.com/clambin/covid19/models"
	"github.com/clambin/simplejson/v3"
	"github.com/clambin/simplejson/v3/dataset"
	"github.com/clambin/simplejson/v3/query"
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
		Query: handler.tableQuery,
	}
}

func (handler *ByCountryHandler) tableQuery(_ context.Context, req query.Request) (response query.Response, err error) {
	var d *dataset.Dataset
	d, err = getStatsByCountry(handler.CovidDB, req.Args, handler.Mode)
	if err != nil {
		return
	}
	return d.GenerateTableResponse(), nil
}

func getStatsByCountry(db covidStore.CovidStore, args query.Args, mode int) (response *dataset.Dataset, err error) {
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
	d := dataset.New()

	for name, entry := range entries {
		if timestamp.IsZero() {
			timestamp = entry.Timestamp
		}

		var value float64
		switch mode {
		case CountryConfirmed:
			value = float64(entry.Confirmed)
		case CountryDeaths:
			value = float64(entry.Deaths)
		}

		d.Add(timestamp, name, value)
	}

	d.FilterByRange(args.Range.From, args.Range.To)

	return d, nil
}
