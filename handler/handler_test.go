package handler_test

import (
	"context"
	"github.com/clambin/covid19/cache"
	mockCovidStore "github.com/clambin/covid19/covid/store/mocks"
	"github.com/clambin/covid19/handler"
	"github.com/clambin/covid19/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

var dbContents = []*models.CountryEntry{
	{
		Timestamp: time.Date(2020, time.November, 1, 0, 0, 0, 0, time.UTC),
		Code:      "A",
		Name:      "A",
		Confirmed: 1,
		Recovered: 0,
		Deaths:    0},
	{
		Timestamp: time.Date(2020, time.November, 2, 0, 0, 0, 0, time.UTC),
		Code:      "B",
		Name:      "B",
		Confirmed: 3,
		Recovered: 0,
		Deaths:    0},
	{
		Timestamp: time.Date(2020, time.November, 2, 0, 0, 0, 0, time.UTC),
		Code:      "A",
		Name:      "A",
		Confirmed: 3,
		Recovered: 1,
		Deaths:    0},
	{
		Timestamp: time.Date(2020, time.November, 4, 0, 0, 0, 0, time.UTC),
		Code:      "B",
		Name:      "B",
		Confirmed: 10,
		Recovered: 5,
		Deaths:    1,
	},
}

func filterByName(input []*models.CountryEntry, name string) (output []*models.CountryEntry) {
	for _, entry := range input {
		if entry.Name == name {
			output = append(output, entry)
		}
	}
	return
}

func TestCovidHandler_Search(t *testing.T) {
	store := &mockCovidStore.CovidStore{}
	c := &cache.Cache{DB: store, Retention: 20 * time.Minute}
	h := handler.Handler{Cache: c}
	targets := h.Search()
	assert.Equal(t, []string{
		"incremental",
		"cumulative",
		"evolution",
		"country-confirmed",
		"country-deaths",
		"country-confirmed-population",
		"country-deaths-population",
	}, targets)

}

func TestCovidHandler_Tags(t *testing.T) {
	store := &mockCovidStore.CovidStore{}
	store.On("GetAllCountryNames").Return([]string{"A", "B"}, nil)

	c := &cache.Cache{DB: store, Retention: 20 * time.Minute}
	h := handler.Handler{Cache: c}

	keys := h.TagKeys(context.Background())
	assert.Equal(t, []string{"Country Name"}, keys)

	values, err := h.TagValues(context.Background(), "Country Name")
	require.NoError(t, err)
	assert.Equal(t, []string{"A", "B"}, values)

	_, err = h.TagValues(context.Background(), "foo")
	assert.Error(t, err)
}
