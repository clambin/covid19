package population_test

import (
	"context"
	"fmt"
	"github.com/clambin/covid19/population"
	"github.com/clambin/go-rapidapi/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestGetPopulation(t *testing.T) {
	rapidMock := &mocks.API{}
	apiClient := population.NewAPIClient("1234")
	apiClient.API = rapidMock

	ctx := context.Background()

	rapidMock.
		On("CallWithContext", mock.Anything, "/population?country_name=Belgium").
		Return([]byte(`{"ok": true, "body": {"country_name": "Belgium", "population": 20}}`), nil).
		Once()

	pop, err := apiClient.GetPopulation(ctx, "Belgium")
	require.NoError(t, err)
	assert.Equal(t, int64(20), pop)

	rapidMock.
		On("CallWithContext", mock.Anything, "/population?country_name=United+States").
		Return([]byte(`{"ok": true, "body": {"country_name": "United States", "population": 40}}`), nil).
		Once()
	pop, err = apiClient.GetPopulation(ctx, "United States")
	require.NoError(t, err)
	assert.Equal(t, int64(40), pop)

	rapidMock.
		On("CallWithContext", mock.Anything, "/population?country_name=%3F%3F").
		Return([]byte(``), fmt.Errorf("404 - Not Found")).
		Once()
	_, err = apiClient.GetPopulation(ctx, "??")
	require.Error(t, err)

	mock.AssertExpectationsForObjects(t, rapidMock)
}

func TestGetCountries(t *testing.T) {
	rapidMock := &mocks.API{}
	apiClient := population.NewAPIClient("1234")
	apiClient.API = rapidMock

	ctx := context.Background()

	rapidMock.
		On("CallWithContext", mock.Anything, "/allcountriesname").
		Return([]byte(`{"ok": true,"body": { "countries": [ "Belgium", "United States" ] }}`), nil).
		Once()

	countries, err := apiClient.GetCountries(ctx)
	assert.NoError(t, err)
	assert.Len(t, countries, 2)
	assert.Contains(t, countries, "Belgium")
	assert.Contains(t, countries, "United States")

	mock.AssertExpectationsForObjects(t, rapidMock)
}
