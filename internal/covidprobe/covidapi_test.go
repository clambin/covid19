package covidprobe

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestGetStats(t *testing.T) {
	apiClient := makeClient()

	response, err := apiClient.getStats()

	assert.Equal(t, nil,         err)
	assert.Equal(t, false,       response.Error)
	assert.Equal(t, 200,         response.StatusCode)
	assert.Equal(t, lastChecked, response.Data.LastChecked)
	assert.Equal(t, 3,           len(response.Data.Covid19Stats))
	assert.Equal(t, lastUpdate , response.Data.Covid19Stats[0].LastUpdate.UTC())
	assert.Equal(t, "A",         response.Data.Covid19Stats[0].Country)
	assert.Equal(t, int64(3),    response.Data.Covid19Stats[0].Confirmed)
	assert.Equal(t, int64(2),    response.Data.Covid19Stats[0].Deaths)
	assert.Equal(t, int64(1),    response.Data.Covid19Stats[0].Recovered)
	assert.Equal(t, lastUpdate , response.Data.Covid19Stats[1].LastUpdate.UTC())
	assert.Equal(t, "B",         response.Data.Covid19Stats[1].Country)
	assert.Equal(t, int64(5),    response.Data.Covid19Stats[1].Confirmed)
	assert.Equal(t, int64(4),    response.Data.Covid19Stats[1].Deaths)
	assert.Equal(t, int64(3),    response.Data.Covid19Stats[1].Recovered)
	assert.Equal(t, lastUpdate , response.Data.Covid19Stats[2].LastUpdate.UTC())
	assert.Equal(t, "B",         response.Data.Covid19Stats[2].Country)
	assert.Equal(t, int64(1),    response.Data.Covid19Stats[2].Confirmed)
	assert.Equal(t, int64(1),    response.Data.Covid19Stats[2].Deaths)
	assert.Equal(t, int64(1),    response.Data.Covid19Stats[2].Recovered)
}

func TestGetCountryStats(t *testing.T) {
	apiClient := makeClient()

	response, err := apiClient.GetCountryStats()

	assert.Equal(t, nil,     err)
	assert.Equal(t, 2,       len(response))
	stats, ok := response["A"]
	assert.Equal(t, true,     ok)
	assert.Equal(t, int64(3), stats.Confirmed)
	assert.Equal(t, int64(2), stats.Deaths)
	assert.Equal(t, int64(1), stats.Recovered)
	stats, ok = response["B"]
	assert.Equal(t, true,     ok)
	assert.Equal(t, int64(6), stats.Confirmed)
	assert.Equal(t, int64(5), stats.Deaths)
	assert.Equal(t, int64(4), stats.Recovered)
}

