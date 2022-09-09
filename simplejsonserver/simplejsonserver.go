package simplejsonserver

import (
	covidStore "github.com/clambin/covid19/db"
	"github.com/clambin/covid19/simplejsonserver/countries"
	"github.com/clambin/covid19/simplejsonserver/evolution"
	"github.com/clambin/covid19/simplejsonserver/mortality"
	"github.com/clambin/covid19/simplejsonserver/summarized"
	"github.com/clambin/covid19/simplejsonserver/updates"
	"github.com/clambin/simplejson/v3"
)

func MakeServer(covidDB covidStore.CovidStore, popDB covidStore.PopulationStore) *simplejson.Server {
	return &simplejson.Server{
		Name: "covid19",
		Handlers: map[string]simplejson.Handler{
			"country-confirmed": countries.ByCountryHandler{
				CovidDB: covidDB,
				Mode:    countries.CountryConfirmed,
			},
			"country-deaths": countries.ByCountryHandler{
				CovidDB: covidDB,
				Mode:    countries.CountryDeaths,
			},
			"country-confirmed-population": countries.ByCountryByPopulationHandler{
				CovidDB: covidDB,
				PopDB:   popDB,
				Mode:    countries.CountryConfirmed,
			},
			"country-deaths-population": countries.ByCountryByPopulationHandler{
				CovidDB: covidDB,
				PopDB:   popDB,
				Mode:    countries.CountryDeaths,
			},
			"country-deaths-vs-confirmed": mortality.Handler{
				CovidDB: covidDB,
			},
			"cumulative": summarized.CumulativeHandler{
				Retriever: summarized.Retriever{DB: covidDB},
			},
			"incremental": summarized.IncrementalHandler{
				Retriever: summarized.Retriever{DB: covidDB},
			},
			"evolution": evolution.Handler{
				CovidDB: covidDB,
			},
			"updates": updates.Handler{
				CovidDB: covidDB,
			},
		},
	}
}
