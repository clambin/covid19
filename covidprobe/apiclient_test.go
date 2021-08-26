package covidprobe_test

import (
	"context"
	"fmt"
	"github.com/clambin/covid19/covidprobe"
	"github.com/clambin/gotools/rapidapi/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestGetCountryStats(t *testing.T) {
	mockAPI := &mocks.API{}
	apiClient := covidprobe.NewAPIClient("1234")
	apiClient.API = mockAPI

	mockAPI.
		On("CallWithContext", mock.AnythingOfType("*context.emptyCtx"), "/v1/stats").
		Return([]byte(goodResponse), nil).
		Once()
	response, err := apiClient.GetCountryStats(context.Background())

	assert.Equal(t, nil, err)
	assert.Equal(t, 2, len(response))
	stats, ok := response["A"]
	assert.Equal(t, true, ok)
	assert.Equal(t, int64(3), stats.Confirmed)
	assert.Equal(t, int64(2), stats.Deaths)
	assert.Equal(t, int64(1), stats.Recovered)
	stats, ok = response["B"]
	assert.Equal(t, true, ok)
	assert.Equal(t, int64(6), stats.Confirmed)
	assert.Equal(t, int64(5), stats.Deaths)
	assert.Equal(t, int64(4), stats.Recovered)

	mockAPI.
		On("CallWithContext", mock.AnythingOfType("*context.emptyCtx"), "/v1/stats").
		Return([]byte(goodResponse), fmt.Errorf("500 - Internal Server Error")).
		Once()
	_, err = apiClient.GetCountryStats(context.Background())

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
					"country": "A",
					"lastUpdate": "2020-12-03T05:28:22+00:00",
					"keyId": "A",
					"confirmed": 3,
					"deaths": 2,
					"recovered": 1
				},
				{
					"city": "B.1",
					"province": null,
					"country": "B",
					"lastUpdate": "2020-12-03T05:28:22+00:00",
					"keyId": "B",
					"confirmed": 5,
					"deaths": 4,
					"recovered": 3
				},
				{
					"city": "B.2",
					"province": null,
					"country": "B",
					"lastUpdate": "2020-12-03T05:28:22+00:00",
					"keyId": "B",
					"confirmed": 1,
					"deaths": 1,
					"recovered": 1
				}
			]
		}
	}`
