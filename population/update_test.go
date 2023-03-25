package population_test

import (
	"context"
	population2 "github.com/clambin/covid19/internal/testtools/db/population"
	"github.com/clambin/covid19/population"
	probeMock "github.com/clambin/covid19/population/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestProbe_Update(t *testing.T) {
	store := &population2.FakeStore{}
	apiClient := probeMock.NewAPIClient(t)

	p := population.New("1234", store)
	p.APIClient = apiClient

	apiClient.On("GetPopulation", mock.Anything, "United States").Return(int64(330), nil)
	apiClient.On("GetPopulation", mock.Anything, "Belgium").Return(int64(11), nil)
	apiClient.On("GetPopulation", mock.Anything, mock.AnythingOfType("string")).Return(int64(0), nil)

	_, err := p.Update(context.Background())
	require.NoError(t, err)

	result, _ := store.List()
	assert.Equal(t, map[string]int64{
		"US": 330,
		"BE": 11,
	}, result)
}
