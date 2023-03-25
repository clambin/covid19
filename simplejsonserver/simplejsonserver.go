package simplejsonserver

import (
	"github.com/clambin/covid19/simplejsonserver/countries"
	"github.com/clambin/covid19/simplejsonserver/evolution"
	"github.com/clambin/covid19/simplejsonserver/mortality"
	"github.com/clambin/covid19/simplejsonserver/summarized"
	"github.com/clambin/covid19/simplejsonserver/updates"
	"github.com/clambin/go-common/httpserver/middleware"
	"github.com/clambin/simplejson/v6"
)

type CovidGetter interface {
	countries.CovidGetter
	mortality.CovidGetter
	summarized.CovidGetter
	evolution.CovidGetter
	updates.CovidGetter
}

type PopulationGetter interface {
	countries.PopulationGetter
}

func New(covidDB CovidGetter, popDB PopulationGetter) *simplejson.Server {
	handlers := map[string]simplejson.Handler{
		"country-confirmed": &countries.ByCountryHandler{
			DB:   covidDB,
			Mode: countries.CountryConfirmed,
		},
		"country-deaths": &countries.ByCountryHandler{
			DB:   covidDB,
			Mode: countries.CountryDeaths,
		},
		"country-confirmed-population": &countries.ByCountryByPopulationHandler{
			CovidDB: covidDB,
			PopDB:   popDB,
			Mode:    countries.CountryConfirmed,
		},
		"country-deaths-population": &countries.ByCountryByPopulationHandler{
			CovidDB: covidDB,
			PopDB:   popDB,
			Mode:    countries.CountryDeaths,
		},
		"country-deaths-vs-confirmed": &mortality.Handler{
			CovidDB: covidDB,
		},
		"cumulative": &summarized.CumulativeHandler{
			Fetcher: summarized.Fetcher{DB: covidDB},
		},
		"incremental": &summarized.IncrementalHandler{
			Fetcher: summarized.Fetcher{DB: covidDB},
		},
		"evolution": &evolution.Handler{
			CovidDB: covidDB,
		},
		"updates": &updates.Handler{
			DB: covidDB,
		},
	}

	return simplejson.New(handlers,
		simplejson.WithQueryMetrics{Name: "covid19"},
		simplejson.WithHTTPMetrics{Option: middleware.PrometheusMetricsOptions{
			Namespace:   "covid",
			Subsystem:   "simplejson",
			Application: "covid19",
		}},
	)
}
