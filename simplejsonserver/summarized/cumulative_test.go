package summarized_test

import (
	"context"
	"fmt"
	"github.com/clambin/covid19/internal/testtools/db/covid"
	"github.com/clambin/covid19/models"
	"github.com/clambin/covid19/simplejsonserver/summarized"
	"github.com/clambin/simplejson/v6"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"strconv"
	"testing"
	"time"
)

func TestCumulativeHandler_Global(t *testing.T) {
	db := covid.FakeStore{Records: dbTotals}
	h := summarized.CumulativeHandler{Fetcher: summarized.Fetcher{DB: &db}}

	args := simplejson.QueryArgs{Args: simplejson.Args{Range: simplejson.Range{To: time.Now()}}}
	ctx := context.Background()

	response, err := h.Endpoints().Query(ctx, simplejson.QueryRequest{QueryArgs: args})
	require.NoError(t, err)
	assert.Equal(t, &simplejson.TableResponse{Columns: []simplejson.Column{
		{Text: "timestamp", Data: simplejson.TimeColumn{
			time.Date(2020, time.November, 1, 0, 0, 0, 0, time.UTC),
			time.Date(2020, time.November, 2, 0, 0, 0, 0, time.UTC),
			time.Date(2020, time.November, 3, 0, 0, 0, 0, time.UTC),
			time.Date(2020, time.November, 4, 0, 0, 0, 0, time.UTC),
		}},
		{Text: "confirmed", Data: simplejson.NumberColumn{1, 3, 3, 10}},
		{Text: "deaths", Data: simplejson.NumberColumn{0, 0, 0, 1}},
	}}, response)
}

func BenchmarkCumulativeHandler_Global(b *testing.B) {
	start := time.Date(2023, time.March, 24, 0, 0, 0, 0, time.UTC)
	end := time.Date(2025, time.March, 24, 0, 0, 0, 0, time.UTC)
	db := buildBigDatabase(start, end)
	h := summarized.CumulativeHandler{Fetcher: summarized.Fetcher{DB: db}}

	args := simplejson.QueryArgs{Args: simplejson.Args{Range: simplejson.Range{
		From: start,
		To:   end,
	}}}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := h.Endpoints().Query(ctx, simplejson.QueryRequest{QueryArgs: args})
		if err != nil {
			b.Fatal(err)
		}
	}
}

func TestCumulativeHandler_Country(t *testing.T) {
	db := covid.FakeStore{Records: dbContents}
	h := summarized.CumulativeHandler{Fetcher: summarized.Fetcher{DB: &db}}

	args := simplejson.QueryArgs{
		Args: simplejson.Args{
			Range: simplejson.Range{
				To: time.Now(),
			},
			AdHocFilters: []simplejson.AdHocFilter{
				{
					Key:      "Country Name",
					Operator: "=",
					Value:    "A",
				},
			},
		},
	}

	ctx := context.Background()

	response, err := h.Endpoints().Query(ctx, simplejson.QueryRequest{QueryArgs: args})
	require.NoError(t, err)
	assert.Equal(t, &simplejson.TableResponse{Columns: []simplejson.Column{
		{Text: "timestamp", Data: simplejson.TimeColumn{time.Date(2020, time.November, 1, 0, 0, 0, 0, time.UTC), time.Date(2020, time.November, 2, 0, 0, 0, 0, time.UTC)}},
		{Text: "confirmed", Data: simplejson.NumberColumn{1, 3}},
		{Text: "deaths", Data: simplejson.NumberColumn{0, 0}},
	}}, response)
}

func TestCumulativeHandler_Tags(t *testing.T) {
	db := covid.FakeStore{Records: dbContents}
	h := summarized.CumulativeHandler{Fetcher: summarized.Fetcher{DB: &db}}
	ctx := context.Background()

	keys := h.Endpoints().TagKeys(ctx)
	assert.Equal(t, []string{"Country Name"}, keys)

	//_, err := h.Endpoints().TagValues(ctx, keys[0])
	//require.Error(t, err)

	values, err := h.Endpoints().TagValues(ctx, keys[0])
	require.NoError(t, err)
	assert.Equal(t, []string{"A", "B"}, values)
}

var (
	dbContents = []models.CountryEntry{
		{
			Timestamp: time.Date(2020, time.November, 1, 0, 0, 0, 0, time.UTC),
			Code:      "A",
			Name:      "A",
			Confirmed: 1,
			Recovered: 0,
			Deaths:    0,
		},
		{
			Timestamp: time.Date(2020, time.November, 2, 0, 0, 0, 0, time.UTC),
			Code:      "B",
			Name:      "B",
			Confirmed: 3,
			Recovered: 0,
			Deaths:    0,
		},
		{
			Timestamp: time.Date(2020, time.November, 2, 0, 0, 0, 0, time.UTC),
			Code:      "A",
			Name:      "A",
			Confirmed: 3,
			Recovered: 1,
			Deaths:    0,
		},
		{
			Timestamp: time.Date(2020, time.November, 4, 0, 0, 0, 0, time.UTC),
			Code:      "B",
			Name:      "B",
			Confirmed: 10,
			Recovered: 5,
			Deaths:    1,
		},
	}

	dbTotals = []models.CountryEntry{
		{
			Timestamp: time.Date(2020, time.November, 1, 0, 0, 0, 0, time.UTC),
			Confirmed: 1,
			Recovered: 0,
			Deaths:    0,
		},
		{
			Timestamp: time.Date(2020, time.November, 2, 0, 0, 0, 0, time.UTC),
			Confirmed: 3,
			Recovered: 0,
			Deaths:    0,
		},
		{
			Timestamp: time.Date(2020, time.November, 3, 0, 0, 0, 0, time.UTC),
			Confirmed: 3,
			Recovered: 1,
			Deaths:    0,
		},
		{
			Timestamp: time.Date(2020, time.November, 4, 0, 0, 0, 0, time.UTC),
			Confirmed: 10,
			Recovered: 5,
			Deaths:    1,
		},
	}
)

var _ summarized.CovidGetter = stubbedStore{}

type stubbedStore struct {
	allForCountry map[string][]models.CountryEntry
	countryNames  []string
	totalsPerDay  []models.CountryEntry
}

func (s stubbedStore) GetAllForCountryName(s2 string) ([]models.CountryEntry, error) {
	results, ok := s.allForCountry[s2]
	if !ok {
		return nil, fmt.Errorf("invalid country: %s", s2)
	}
	return results, nil
}

func (s stubbedStore) GetAllCountryNames() ([]string, error) {
	return s.countryNames, nil
}

func (s stubbedStore) GetTotalsPerDay() ([]models.CountryEntry, error) {
	return s.totalsPerDay, nil
}

func buildBigDatabase(from, to time.Time) stubbedStore {
	allForCountry := make(map[string][]models.CountryEntry)
	countryNames := make([]string, 0, 193)
	var totalsForDay []models.CountryEntry

	for country := 0; country < 193; country++ {
		content := make([]models.CountryEntry, 0, 193)
		countryName := strconv.Itoa(country)
		for timestamp := from; !timestamp.After(to); timestamp = timestamp.Add(24 * time.Hour) {
			content = append(content, models.CountryEntry{
				Timestamp: timestamp,
				Code:      countryName,
				Name:      countryName,
			})
			if country == 0 {
				totalsForDay = append(totalsForDay, models.CountryEntry{Timestamp: timestamp})
			}
		}
		allForCountry[countryName] = content
		if country == 0 {
			countryNames = append(countryNames, countryName)
		}
	}
	return stubbedStore{
		allForCountry: allForCountry,
		countryNames:  countryNames,
		totalsPerDay:  totalsForDay,
	}
}
