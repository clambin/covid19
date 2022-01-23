package summarized_test

import (
	"context"
	"errors"
	"github.com/clambin/covid19/cache"
	mockCovidStore "github.com/clambin/covid19/covid/store/mocks"
	"github.com/clambin/covid19/models"
	"github.com/clambin/covid19/simplejsonserver/summarized"
	"github.com/clambin/simplejson"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"strconv"
	"testing"
	"time"
)

func TestCumulativeHandler_Global(t *testing.T) {
	dbh := &mockCovidStore.CovidStore{}
	dbh.On("GetAll").Return(dbContents, nil)

	c := &cache.Cache{DB: dbh, Retention: 20 * time.Minute}
	h := summarized.CumulativeHandler{Cache: c}

	args := simplejson.TableQueryArgs{
		Args: simplejson.Args{
			Range: simplejson.Range{
				To: time.Now(),
			},
		},
	}

	ctx := context.Background()

	response, err := h.Endpoints().TableQuery(ctx, &args)
	require.NoError(t, err)
	require.Len(t, response.Columns, 3)
	for i := 0; i < 3; i++ {
		require.Len(t, response.Columns[i].Data, 3)
	}
	assert.Equal(t, simplejson.TableQueryResponseNumberColumn{0, 0, 1}, response.Columns[1].Data)
	assert.Equal(t, simplejson.TableQueryResponseNumberColumn{1, 6, 13}, response.Columns[2].Data)

	mock.AssertExpectationsForObjects(t, dbh)
}

func BenchmarkCumulativeHandler_Global(b *testing.B) {
	from, to, bigContents := buildBigDatabase()
	dbh := &mockCovidStore.CovidStore{}
	dbh.On("GetAll").Return(bigContents, nil)

	c := &cache.Cache{DB: dbh, Retention: 20 * time.Minute}
	h := summarized.CumulativeHandler{Cache: c}

	args := simplejson.TableQueryArgs{
		Args: simplejson.Args{
			Range: simplejson.Range{
				From: from,
				To:   to,
			},
		},
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := h.Endpoints().TableQuery(ctx, &args)
		if err != nil {
			b.Fatal(err)
		}
	}
	mock.AssertExpectationsForObjects(b, dbh)
}

func buildBigDatabase() (from, to time.Time, content []models.CountryEntry) {
	from = time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC)
	timestamp := from
	for i := 0; i < 2*365; i++ {
		to = timestamp
		for c := 0; c < 193; c++ {
			country := strconv.Itoa(c)
			content = append(content, models.CountryEntry{
				Timestamp: timestamp,
				Code:      country,
				Name:      country,
			})
		}
		timestamp = timestamp.Add(24 * time.Hour)
	}
	return
}

func TestCumulativeHandler_Country(t *testing.T) {
	dbh := &mockCovidStore.CovidStore{}
	dbh.On("GetAllForCountryName", "A").Return(filterByName(dbContents, "A"), nil)

	c := &cache.Cache{DB: dbh, Retention: 20 * time.Minute}
	h := summarized.CumulativeHandler{Cache: c}

	args := simplejson.TableQueryArgs{
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

	response, err := h.Endpoints().TableQuery(ctx, &args)
	require.NoError(t, err)
	require.Len(t, response.Columns, 3)
	for i := 0; i < 3; i++ {
		require.Len(t, response.Columns[i].Data, 2)
	}
	assert.Equal(t, simplejson.TableQueryResponseNumberColumn{0, 0}, response.Columns[1].Data)
	assert.Equal(t, simplejson.TableQueryResponseNumberColumn{1, 3}, response.Columns[2].Data)

	mock.AssertExpectationsForObjects(t, dbh)
}

func TestCumulativeHandler_Tags(t *testing.T) {
	dbh := &mockCovidStore.CovidStore{}
	c := &cache.Cache{DB: dbh, Retention: 20 * time.Minute}
	h := summarized.CumulativeHandler{Cache: c}

	ctx := context.Background()

	keys := h.Endpoints().TagKeys(ctx)
	assert.Equal(t, []string{"Country Name"}, keys)

	dbh.On("GetAllCountryNames").Return(nil, errors.New("db error")).Once()
	_, err := h.Endpoints().TagValues(ctx, keys[0])
	require.Error(t, err)

	dbh.On("GetAllCountryNames").Return([]string{"A", "B"}, nil)
	var values []string
	values, err = h.Endpoints().TagValues(ctx, keys[0])
	require.NoError(t, err)
	assert.Equal(t, []string{"A", "B"}, values)

	mock.AssertExpectationsForObjects(t, dbh)
}

var dbContents = []models.CountryEntry{
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

func filterByName(input []models.CountryEntry, name string) (output []models.CountryEntry) {
	for _, entry := range input {
		if entry.Name == name {
			output = append(output, entry)
		}
	}
	return
}
