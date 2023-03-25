package fetcher_test

import (
	"context"
	"errors"
	"github.com/clambin/covid19/covid/fetcher"
	"github.com/clambin/covid19/models"
	"github.com/clambin/go-rapidapi/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"sort"
	"testing"
	"time"
)

func TestClient_Fetch(t *testing.T) {
	tests := []struct {
		name         string
		responseBody []byte
		responseErr  error
		wantErr      assert.ErrorAssertionFunc
		want         []models.CountryEntry
	}{
		{
			name:         "valid",
			responseBody: []byte(goodResponse),
			responseErr:  nil,
			wantErr:      assert.NoError,
			want: []models.CountryEntry{
				{Timestamp: time.Date(2020, time.December, 3, 5, 28, 22, 0, time.UTC), Code: "", Name: "Belgium", Confirmed: 3, Recovered: 1, Deaths: 2},
				{Timestamp: time.Date(2020, time.December, 3, 5, 28, 22, 0, time.UTC), Code: "", Name: "US", Confirmed: 6, Recovered: 4, Deaths: 5},
				{Timestamp: time.Date(2020, time.December, 3, 5, 28, 22, 0, time.UTC), Code: "", Name: "invalid_country", Confirmed: 1, Recovered: 1, Deaths: 1},
			},
		},
		{
			name:         "call fails",
			responseBody: []byte(""),
			responseErr:  errors.New("fail"),
			wantErr:      assert.Error,
			want:         nil,
		},
		{
			name:         "invalid data",
			responseBody: []byte("invalid data"),
			responseErr:  errors.New("fail"),
			wantErr:      assert.Error,
			want:         nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAPI := &mocks.API{}
			client := fetcher.Client{API: mockAPI}

			mockAPI.
				On("CallWithContext", mock.AnythingOfType("*context.emptyCtx"), "/v1/stats").
				Return(tt.responseBody, tt.responseErr).
				Once()

			response, err := client.Fetch(context.Background())
			tt.wantErr(t, err)
			sort.Slice(response, func(i, j int) bool {
				// API only returns entries for the current day, so no need to sort on date
				return response[i].Name < response[j].Name
			})
			assert.Equal(t, tt.want, response)
		})
	}
}

const goodResponse = `
	{
		"error": false,
		"statusCode": 200,
		"message": "OK",
		"data": {
			"lastChecked": "2020-12-03T11:23:52.193Z",
			"covid19Stats": [
				{
					"city": null,
					"province": null,
					"country": "Belgium",
					"lastUpdate": "2020-12-03T05:28:22+00:00",
					"keyId": "A",
					"confirmed": 3,
					"deaths": 2,
					"recovered": 1
				},
				{
					"city": "B.1",
					"province": null,
					"country": "US",
					"lastUpdate": "2020-12-03T05:28:22+00:00",
					"keyId": "B",
					"confirmed": 5,
					"deaths": 4,
					"recovered": 3
				},
				{
					"city": "B.2",
					"province": null,
					"country": "US",
					"lastUpdate": "2020-12-03T05:28:22+00:00",
					"keyId": "B",
					"confirmed": 1,
					"deaths": 1,
					"recovered": 1
				},
				{
					"city": "C.1",
					"province": null,
					"country": "invalid_country",
					"lastUpdate": "2020-12-03T05:28:22+00:00",
					"keyId": "C",
					"confirmed": 1,
					"deaths": 1,
					"recovered": 1
				}
			]
		}
	}`
