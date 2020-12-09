package covid_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetCountryStats(t *testing.T) {
	apiClient := makeClient()

	response, err := apiClient.GetCountryStats()

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
}
