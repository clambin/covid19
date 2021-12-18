package handler_test

import (
	"context"
	"github.com/clambin/covid19/cache"
	"github.com/clambin/covid19/covid/probe/fetcher"
	mockCovidStore "github.com/clambin/covid19/covid/store/mocks"
	"github.com/clambin/covid19/handler"
	"github.com/clambin/covid19/models"
	grafanaJson "github.com/clambin/grafana-json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestEvolution(t *testing.T) {
	dbh := &mockCovidStore.CovidStore{}
	dbh.
		On(
			"GetAllForRange",
			dbContents[len(dbContents)-1].Timestamp.Add(-7*24*time.Hour),
			dbContents[len(dbContents)-1].Timestamp,
		).
		Return(dbContents, nil)

	c := &cache.Cache{DB: dbh, Retention: 20 * time.Minute}
	h := handler.Handler{Cache: c}

	args := grafanaJson.TableQueryArgs{
		CommonQueryArgs: grafanaJson.CommonQueryArgs{
			Range: grafanaJson.QueryRequestRange{
				To: dbContents[len(dbContents)-1].Timestamp,
			},
		},
	}

	ctx := context.Background()

	response, err := h.TableQuery(ctx, "evolution", &args)
	require.NoError(t, err)
	require.Len(t, response.Columns, 3)
	assert.Equal(t, "timestamp", response.Columns[0].Text)
	assert.Len(t, response.Columns[0].Data, 2)
	assert.Equal(t, grafanaJson.TableQueryResponseColumn{
		Text: "country",
		Data: grafanaJson.TableQueryResponseStringColumn{"A", "B"},
	}, response.Columns[1])
	assert.Equal(t, grafanaJson.TableQueryResponseColumn{
		Text: "increase",
		Data: grafanaJson.TableQueryResponseNumberColumn{1.0, 3.5},
	}, response.Columns[2])

	mock.AssertExpectationsForObjects(t, dbh)
}

func TestEvolution_NoEndDate(t *testing.T) {
	dbh := &mockCovidStore.CovidStore{}
	c := &cache.Cache{DB: dbh, Retention: 20 * time.Minute}
	h := handler.Handler{Cache: c}

	args := grafanaJson.TableQueryArgs{}
	ctx := context.Background()

	dbContents2 := []models.CountryEntry{
		{
			Timestamp: time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC),
			Code:      "A",
			Name:      "A",
			Confirmed: 1,
			Recovered: 0,
			Deaths:    0,
		},
	}
	dbContents2 = append(dbContents2, dbContents...)
	dbh.
		On("GetAll").
		Return(dbContents2, nil)

	response, err := h.TableQuery(ctx, "evolution", &args)
	require.NoError(t, err)
	require.Len(t, response.Columns, 3)
	assert.Equal(t, "timestamp", response.Columns[0].Text)
	assert.Len(t, response.Columns[0].Data, 2)
	assert.Equal(t, grafanaJson.TableQueryResponseColumn{
		Text: "country",
		Data: grafanaJson.TableQueryResponseStringColumn{"A", "B"},
	}, response.Columns[1])
	assert.Equal(t, grafanaJson.TableQueryResponseColumn{
		Text: "increase",
		Data: grafanaJson.TableQueryResponseNumberColumn{1.0, 3.5},
	}, response.Columns[2])

	mock.AssertExpectationsForObjects(t, dbh)
}

func TestEvolution_NoData(t *testing.T) {
	dbh := &mockCovidStore.CovidStore{}
	dbh.
		On(
			"GetAllForRange",
			time.Date(2020, time.October, 24, 0, 0, 0, 0, time.UTC),
			time.Date(2020, time.October, 31, 0, 0, 0, 0, time.UTC),
		).
		Return([]models.CountryEntry{}, nil)

	c := &cache.Cache{DB: dbh, Retention: 20 * time.Minute}
	h := handler.Handler{Cache: c}

	args := grafanaJson.TableQueryArgs{
		CommonQueryArgs: grafanaJson.CommonQueryArgs{
			Range: grafanaJson.QueryRequestRange{
				To: time.Date(2020, time.October, 31, 0, 0, 0, 0, time.UTC),
			},
		},
	}
	ctx := context.Background()

	response, err := h.TableQuery(ctx, "evolution", &args)
	require.NoError(t, err)
	require.Len(t, response.Columns, 3)
	for _, column := range response.Columns {
		assert.Empty(t, column.Data)
	}

	mock.AssertExpectationsForObjects(t, dbh)
}

func TestMortalityVsConfirmed(t *testing.T) {
	dbh := &mockCovidStore.CovidStore{}
	dbh.
		On("GetAllCountryNames").
		Return([]string{"AA", "BB"}, nil)
	dbh.
		On("GetLatestForCountriesByTime", []string{"AA", "BB"}, mock.AnythingOfType("time.Time")).
		Return(map[string]models.CountryEntry{
			"AA": {
				Timestamp: time.Date(2021, 12, 17, 0, 0, 0, 0, time.UTC),
				Code:      "A",
				Name:      "AA",
				Confirmed: 100,
				Deaths:    10,
			},
			"BB": {
				Timestamp: time.Date(2021, 12, 17, 0, 0, 0, 0, time.UTC),
				Code:      "B",
				Name:      "BB",
				Confirmed: 200,
				Deaths:    10,
			},
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

	response, err := h.TableQuery(ctx, "country-deaths-vs-confirmed", &args)
	require.NoError(t, err)
	require.Len(t, response.Columns, 3)
	assert.Equal(t, "timestamp", response.Columns[0].Text)
	assert.Len(t, response.Columns[0].Data, 2)
	assert.Equal(t, grafanaJson.TableQueryResponseColumn{
		Text: "country",
		Data: grafanaJson.TableQueryResponseStringColumn{"A", "B"},
	}, response.Columns[1])
	assert.Equal(t, grafanaJson.TableQueryResponseColumn{
		Text: "ratio",
		Data: grafanaJson.TableQueryResponseNumberColumn{0.1, 0.05},
	}, response.Columns[2])

	mock.AssertExpectationsForObjects(t, dbh)
}

func BenchmarkHandler_TableQuery_Evolution(b *testing.B) {
	var bigData []models.CountryEntry
	timestamp := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	for i := 0; i < 2*365; i++ {
		for name, code := range fetcher.CountryCodes {
			bigData = append(bigData, models.CountryEntry{
				Timestamp: timestamp,
				Code:      code,
				Name:      name,
			})
		}
		timestamp.Add(24 * time.Hour)
	}

	dbh := &mockCovidStore.CovidStore{}
	dbh.On("GetAllForRange", timestamp.Add(-7*24*time.Hour), timestamp).Return(bigData, nil)

	c := &cache.Cache{DB: dbh, Retention: 20 * time.Minute}
	h := handler.Handler{Cache: c}

	args := grafanaJson.TableQueryArgs{
		CommonQueryArgs: grafanaJson.CommonQueryArgs{
			Range: grafanaJson.QueryRequestRange{
				To: timestamp,
			},
		},
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < 100; i++ {
		_, err := h.TableQuery(ctx, "evolution", &args)
		if err != nil {
			panic(err)
		}
	}

}
