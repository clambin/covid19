package population

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestGetPopulation(t *testing.T) {
	apiClient := makeClient()

	response, err := apiClient.GetPopulation()

	assert.Equal(t, nil,              err)
	assert.Equal(t, 2,                len(response))
	population, ok := response["BE"]
	assert.Equal(t, true,      ok)
	assert.Equal(t, int64(11248330),  population)
	population, ok = response["US"]
	assert.Equal(t, true,             ok)
	assert.Equal(t, int64(321645000), population)
	population, ok = response["??"]
	assert.Equal(t, false,            ok)
}

