package backfill_test

import (
	"github.com/clambin/covid19/backfill"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestClient_GetCountries(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(covidAPI))
	defer s.Close()
	c := backfill.Client{URL: s.URL}

	countries, err := c.GetCountries()
	require.NoError(t, err)
	assert.Equal(t, backfill.Countries{
		"belgium": backfill.Country{Name: "Belgium", Code: "BE"},
		"myanmar": backfill.Country{Name: "Myanmar", Code: "MM"},
	}, countries)
}

func TestClient_GetHistoricalData(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(covidAPI))
	defer s.Close()
	c := backfill.Client{URL: s.URL}

	tests := []struct {
		name    string
		slug    string
		want    []backfill.CountryData
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "valid",
			slug: "belgium",
			want: []backfill.CountryData{
				{Date: time.Date(2020, time.January, 22, 0, 0, 0, 0, time.UTC), Confirmed: 0, Recovered: 0, Deaths: 0},
				{Date: time.Date(2020, time.February, 4, 0, 0, 0, 0, time.UTC), Confirmed: 1, Recovered: 0, Deaths: 0},
			},
			wantErr: assert.NoError,
		},
		{
			name: "valid",
			slug: "myanmar",
			want: []backfill.CountryData{
				{Date: time.Date(2020, time.January, 31, 0, 0, 0, 0, time.UTC), Confirmed: 8, Recovered: 0, Deaths: 0},
			},
			wantErr: assert.NoError,
		},
		{
			name:    "missing",
			slug:    "foobar",
			want:    nil,
			wantErr: assert.Error,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := c.GetHistoricalData(tt.slug)
			tt.wantErr(t, err)
			assert.Equal(t, tt.want, data)
		})
	}
}

func TestClient_GetHistoricalData_TooManyRequests(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "", http.StatusTooManyRequests)
	}))
	defer s.Close()
	c := backfill.Client{URL: s.URL}

	backfill.MaxRetries = 2
	_, err := c.GetHistoricalData("belgium")
	assert.Error(t, err)
}
