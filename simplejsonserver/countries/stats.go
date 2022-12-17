package countries

import (
	"fmt"
	covidStore "github.com/clambin/covid19/db"
	"github.com/clambin/covid19/models"
	"github.com/clambin/simplejson/v5"
	"github.com/clambin/simplejson/v5/pkg/data"
	"sort"
	"time"
)

func getStatsByCountry(db covidStore.CovidStore, args simplejson.QueryArgs, mode int) (*data.Table, error) {
	var err error
	var entries map[string]models.CountryEntry

	if args.Range.To.IsZero() {
		entries, err = db.GetLatestForCountries()
	} else {
		entries, err = db.GetLatestForCountriesByTime(args.Range.To)
	}

	if err != nil {
		return nil, fmt.Errorf("database: %w", err)
	}

	var timestamp time.Time
	columns := make([]data.Column, 0, len(entries))
	countries := make([]string, 0, len(entries))

	for countryName := range entries {
		countries = append(countries, countryName)
	}
	sort.Strings(countries)

	for _, name := range countries {
		entry := entries[name]

		if timestamp.IsZero() {
			timestamp = entry.Timestamp
			columns = append(columns, data.Column{Name: "timestamp", Values: []time.Time{timestamp}})
		}

		var value float64
		switch mode {
		case CountryConfirmed:
			value = float64(entry.Confirmed)
		case CountryDeaths:
			value = float64(entry.Deaths)
		}

		columns = append(columns, data.Column{Name: name, Values: []float64{value}})
	}

	return data.New(columns...).Filter(args.Args), nil
}
