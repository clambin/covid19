package handler_test

import (
	"context"
	"github.com/clambin/covid19/cache"
	mockCovidStore "github.com/clambin/covid19/covid/store/mocks"
	"github.com/clambin/covid19/handler"
	"github.com/clambin/covid19/models"
	mockPopulationStore "github.com/clambin/covid19/population/store/mocks"
	grafanaJson "github.com/clambin/grafana-json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestConfirmedByCountry(t *testing.T) {
	dbh := &mockCovidStore.CovidStore{}
	dbh.On("GetAllCountryNames").Return([]string{"A", "B"}, nil)
	dbh.On("GetLatestForCountriesByTime", []string{"A", "B"}, mock.AnythingOfType("time.Time")).Return(map[string]*models.CountryEntry{
		"A": {Timestamp: time.Now(), Confirmed: 3},
		"B": {Timestamp: time.Now(), Confirmed: 10},
	}, nil)

	c := &cache.Cache{DB: dbh, Retention: 20 * time.Minute}
	h := handler.Handler{Cache: c}

	args := grafanaJson.TableQueryArgs{
		CommonQueryArgs: grafanaJson.CommonQueryArgs{
			Range: grafanaJson.QueryRequestRange{
				To: time.Now(),
			},
		},
	}

	ctx := context.Background()

	response, err := h.TableQuery(ctx, "country-confirmed", &args)
	require.NoError(t, err)
	require.Len(t, response.Columns, 3)
	assert.Equal(t, "timestamp", response.Columns[0].Text)
	assert.Len(t, response.Columns[0].Data, 1)
	assert.Equal(t, grafanaJson.TableQueryResponseColumn{
		Text: "A",
		Data: grafanaJson.TableQueryResponseNumberColumn{3.0},
	}, response.Columns[1])
	assert.Equal(t, grafanaJson.TableQueryResponseColumn{
		Text: "B",
		Data: grafanaJson.TableQueryResponseNumberColumn{10.0},
	}, response.Columns[2])

	mock.AssertExpectationsForObjects(t, dbh)
}

func TestDeathsByCountry(t *testing.T) {
	dbh := &mockCovidStore.CovidStore{}
	dbh.On("GetAllCountryNames").Return([]string{"A", "B"}, nil)
	dbh.On("GetLatestForCountriesByTime", []string{"A", "B"}, mock.AnythingOfType("time.Time")).Return(map[string]*models.CountryEntry{
		"A": {Timestamp: time.Now(), Deaths: 0},
		"B": {Timestamp: time.Now(), Deaths: 1},
	}, nil)

	c := &cache.Cache{DB: dbh, Retention: 20 * time.Minute}
	h := handler.Handler{Cache: c}

	args := grafanaJson.TableQueryArgs{
		CommonQueryArgs: grafanaJson.CommonQueryArgs{
			Range: grafanaJson.QueryRequestRange{
				To: time.Now(),
			},
		},
	}

	ctx := context.Background()

	response, err := h.TableQuery(ctx, "country-deaths", &args)
	require.NoError(t, err)
	assert.Equal(t, "timestamp", response.Columns[0].Text)
	assert.Len(t, response.Columns[0].Data, 1)
	assert.Equal(t, grafanaJson.TableQueryResponseColumn{
		Text: "A",
		Data: grafanaJson.TableQueryResponseNumberColumn{0.0},
	}, response.Columns[1])
	assert.Equal(t, grafanaJson.TableQueryResponseColumn{
		Text: "B",
		Data: grafanaJson.TableQueryResponseNumberColumn{1.0},
	}, response.Columns[2])

	mock.AssertExpectationsForObjects(t, dbh)
}

func TestConfirmedByCountryByPopulation(t *testing.T) {
	dbh := &mockCovidStore.CovidStore{}
	dbh.
		On("GetAllCountryNames").
		Return([]string{"Belgium", "US"}, nil)
	dbh.
		On("GetLatestForCountriesByTime", []string{"Belgium", "US"}, time.Date(2020, 1, 3, 0, 0, 0, 0, time.UTC)).
		Return(map[string]*models.CountryEntry{
			"Belgium": {Timestamp: time.Date(2020, 1, 3, 0, 0, 0, 0, time.UTC), Name: "Belgium", Code: "BE", Confirmed: 200},
			"US":      {Timestamp: time.Date(2020, 1, 3, 0, 0, 0, 0, time.UTC), Name: "US", Code: "US", Confirmed: 200},
		}, nil)

	dbh2 := &mockPopulationStore.PopulationStore{}
	dbh2.On("List").Return(map[string]int64{
		"BE": 10,
		"US": 20,
	}, nil)

	c := &cache.Cache{DB: dbh, Retention: 20 * time.Minute}
	h := handler.Handler{Cache: c, PopulationStore: dbh2}

	args := grafanaJson.TableQueryArgs{
		CommonQueryArgs: grafanaJson.CommonQueryArgs{
			Range: grafanaJson.QueryRequestRange{
				To: time.Date(2020, 1, 3, 0, 0, 0, 0, time.UTC),
			},
		},
	}

	ctx := context.Background()

	response, err := h.TableQuery(ctx, "country-confirmed-population", &args)
	require.NoError(t, err)
	require.Len(t, response.Columns, 3)
	assert.Equal(t, "timestamp", response.Columns[0].Text)
	assert.Len(t, response.Columns[0].Data, 2)
	assert.Equal(t, grafanaJson.TableQueryResponseColumn{
		Text: "country",
		Data: grafanaJson.TableQueryResponseStringColumn{"BE", "US"},
	}, response.Columns[1])
	assert.Equal(t, grafanaJson.TableQueryResponseColumn{
		Text: "confirmed",
		Data: grafanaJson.TableQueryResponseNumberColumn{20.0, 10.0},
	}, response.Columns[2])

	mock.AssertExpectationsForObjects(t, dbh)
}

func TestDeathsByCountryByPopulation(t *testing.T) {
	dbh := &mockCovidStore.CovidStore{}
	dbh.
		On("GetAllCountryNames").
		Return([]string{"Belgium", "US"}, nil)
	dbh.
		On("GetLatestForCountriesByTime", []string{"Belgium", "US"}, time.Date(2020, 1, 3, 0, 0, 0, 0, time.UTC)).
		Return(map[string]*models.CountryEntry{
			"Belgium": {Timestamp: time.Date(2020, 1, 3, 0, 0, 0, 0, time.UTC), Name: "Belgium", Code: "BE", Deaths: 200},
			"US":      {Timestamp: time.Date(2020, 1, 3, 0, 0, 0, 0, time.UTC), Name: "US", Code: "US", Deaths: 200},
		}, nil)

	dbh2 := &mockPopulationStore.PopulationStore{}
	dbh2.On("List").Return(map[string]int64{
		"BE": 10,
		"US": 20,
	}, nil)

	c := &cache.Cache{DB: dbh, Retention: 20 * time.Minute}
	h := handler.Handler{Cache: c, PopulationStore: dbh2}

	args := grafanaJson.TableQueryArgs{
		CommonQueryArgs: grafanaJson.CommonQueryArgs{
			Range: grafanaJson.QueryRequestRange{
				To: time.Date(2020, 1, 3, 0, 0, 0, 0, time.UTC),
			},
		},
	}

	ctx := context.Background()

	response, err := h.TableQuery(ctx, "country-deaths-population", &args)
	require.NoError(t, err)
	require.Len(t, response.Columns, 3)
	assert.Equal(t, "timestamp", response.Columns[0].Text)
	assert.Len(t, response.Columns[0].Data, 2)
	assert.Equal(t, grafanaJson.TableQueryResponseColumn{
		Text: "country",
		Data: grafanaJson.TableQueryResponseStringColumn{"BE", "US"},
	}, response.Columns[1])
	assert.Equal(t, grafanaJson.TableQueryResponseColumn{
		Text: "deaths",
		Data: grafanaJson.TableQueryResponseNumberColumn{20.0, 10.0},
	}, response.Columns[2])

	mock.AssertExpectationsForObjects(t, dbh)
}
