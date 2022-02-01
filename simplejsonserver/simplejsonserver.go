package simplejsonserver

import (
	"github.com/clambin/covid19/cache"
	covidStore "github.com/clambin/covid19/covid/store"
	populationStore "github.com/clambin/covid19/population/store"
	"github.com/clambin/covid19/simplejsonserver/countries"
	"github.com/clambin/covid19/simplejsonserver/evolution"
	"github.com/clambin/covid19/simplejsonserver/mortality"
	"github.com/clambin/covid19/simplejsonserver/summarized"
	"github.com/clambin/covid19/simplejsonserver/updates"
	"github.com/clambin/simplejson/v2"
)

// 	"incremental",
//	"cumulative",
//	"evolution",
//	"updates",

func MakeServer(covidDB covidStore.CovidStore, popDB populationStore.PopulationStore, dbCache *cache.Cache) *simplejson.Server {
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
				Cache: dbCache,
			},
			"incremental": summarized.IncrementalHandler{
				Cache: dbCache,
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
