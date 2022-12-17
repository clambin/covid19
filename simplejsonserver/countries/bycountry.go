package countries

import (
	"context"
	covidStore "github.com/clambin/covid19/db"
	"github.com/clambin/simplejson/v5"
	"github.com/clambin/simplejson/v5/pkg/data"
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

func (handler *ByCountryHandler) Endpoints() (endpoints simplejson.Endpoints) {
	return simplejson.Endpoints{
		Query: handler.tableQuery,
	}
}

func (handler *ByCountryHandler) tableQuery(_ context.Context, req simplejson.QueryRequest) (response simplejson.Response, err error) {
	var d *data.Table
	d, err = getStatsByCountry(handler.CovidDB, req.QueryArgs, handler.Mode)
	if err != nil {
		return
	}
	return d.CreateTableResponse(), nil
}
