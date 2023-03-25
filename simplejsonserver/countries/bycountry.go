package countries

import (
	"context"
	"github.com/clambin/covid19/models"
	"github.com/clambin/simplejson/v6"
	"github.com/clambin/simplejson/v6/pkg/data"
	"time"
)

const (
	CountryConfirmed = iota
	CountryDeaths
)

// ByCountryHandler returns the latest stats by country
type ByCountryHandler struct {
	DB   CovidGetter
	Mode int
}

type CovidGetter interface {
	GetLatestForCountries(time time.Time) (map[string]models.CountryEntry, error)
}

var _ simplejson.Handler = &ByCountryHandler{}

func (handler *ByCountryHandler) Endpoints() (endpoints simplejson.Endpoints) {
	return simplejson.Endpoints{
		Query: handler.tableQuery,
	}
}

func (handler *ByCountryHandler) tableQuery(_ context.Context, req simplejson.QueryRequest) (response simplejson.Response, err error) {
	var d *data.Table
	d, err = getStatsByCountry(handler.DB, req.QueryArgs, handler.Mode)
	if err != nil {
		return
	}
	return d.CreateTableResponse(), nil
}
