package population_test

import (
	"context"
	popDBMock "github.com/clambin/covid19/db/mocks"
	"github.com/clambin/covid19/population"
	probeMock "github.com/clambin/covid19/population/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestProbe_Update(t *testing.T) {
	store := &popDBMock.PopulationStore{}
	apiClient := &probeMock.APIClient{}

	p := population.New("1234", store)
	p.APIClient = apiClient

	apiClient.On("GetPopulation", mock.Anything, "United States").Return(int64(330), nil)
	apiClient.On("GetPopulation", mock.Anything, "Belgium").Return(int64(11), nil)
	apiClient.On("GetPopulation", mock.Anything, mock.AnythingOfType("string")).Return(int64(0), nil)
	store.On("Add", "US", int64(330)).Return(nil)
	store.On("Add", "BE", int64(11)).Return(nil)

	err := p.Update(context.Background())
	require.NoError(t, err)

	mock.AssertExpectationsForObjects(t, store, apiClient)
}
