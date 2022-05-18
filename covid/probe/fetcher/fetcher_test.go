package fetcher_test

import (
	"context"
	"fmt"
	"github.com/clambin/covid19/covid/probe/fetcher"
	"github.com/clambin/go-rapidapi/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestGetCountryStats(t *testing.T) {
	mockAPI := &mocks.API{}
	client := fetcher.Client{API: mockAPI}

	mockAPI.
		On("CallWithContext", mock.AnythingOfType("*context.emptyCtx"), "/v1/stats").
		Return([]byte(goodResponse), nil).
		Once()

	response, err := client.GetCountryStats(context.Background())
	require.NoError(t, err)
	require.Len(t, response, 2)

	indexBE := 0
	indexUS := 1
	if response[0].Code == "US" {
		indexBE = 1
		indexUS = 0
	}

	assert.Equal(t, "BE", response[indexBE].Code)
	assert.Equal(t, "Belgium", response[indexBE].Name)
	assert.Equal(t, int64(3), response[indexBE].Confirmed)
	assert.Equal(t, int64(2), response[indexBE].Deaths)
	assert.Equal(t, int64(1), response[indexBE].Recovered)

	assert.Equal(t, "US", response[indexUS].Code)
	assert.Equal(t, "US", response[indexUS].Name)
	assert.Equal(t, int64(6), response[indexUS].Confirmed)
	assert.Equal(t, int64(5), response[indexUS].Deaths)
	assert.Equal(t, int64(4), response[indexUS].Recovered)

	mockAPI.
		On("CallWithContext", mock.AnythingOfType("*context.emptyCtx"), "/v1/stats").
		Return([]byte(``), fmt.Errorf("500 - Internal Server Error")).
		Once()

	_, err = client.GetCountryStats(context.Background())
	require.Error(t, err)
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
