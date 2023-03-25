package countries

import (
	"fmt"
	"github.com/clambin/simplejson/v6"
	"github.com/clambin/simplejson/v6/pkg/data"
	"sort"
	"time"
)

func getStatsByCountry(db CovidGetter, args simplejson.QueryArgs, mode int) (*data.Table, error) {
	entries, err := db.GetLatestForCountries(args.Range.To)
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
