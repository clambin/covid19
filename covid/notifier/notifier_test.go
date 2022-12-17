package notifier_test

import (
	"fmt"
	"github.com/clambin/covid19/covid/notifier"
	mockRouter "github.com/clambin/covid19/covid/notifier/mocks"
	mockStore "github.com/clambin/covid19/db/mocks"
	"github.com/clambin/covid19/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestNotifier_Notify(t *testing.T) {
	db := mockStore.NewCovidStore(t)
	r := mockRouter.NewRouter(t)

	timestamp := time.Now()
	db.
		On("GetLatestForCountries").
		Return(map[string]models.CountryEntry{"Belgium": {Name: "Belgium", Code: "BE", Timestamp: timestamp, Confirmed: 5, Recovered: 1, Deaths: 0}}, nil).
		Once()

	n, err := notifier.NewNotifier(r, []string{"Belgium"}, db)
	require.NoError(t, err)

	err = n.Notify([]models.CountryEntry{})
	assert.NoError(t, err)

	err = n.Notify([]models.CountryEntry{
		{Name: "Belgium", Code: "BE", Timestamp: timestamp.Add(24 * time.Hour), Confirmed: 5, Recovered: 1, Deaths: 0},
	})
	assert.NoError(t, err)

	r.On("Send", "New probe data for Belgium", "Confirmed: 5, deaths: 1, recovered: 4").Return(nil).Once()
	err = n.Notify([]models.CountryEntry{
		{Name: "Belgium", Code: "BE", Timestamp: timestamp.Add(24 * time.Hour), Confirmed: 10, Recovered: 5, Deaths: 1},
		{Name: "US", Code: "US", Timestamp: timestamp, Confirmed: 50, Recovered: 10, Deaths: 5},
	})
	assert.NoError(t, err)

	r.On("Send", "New probe data for Belgium", "Confirmed: 5, deaths: 1, recovered: 3").Return(nil).Once()
	err = n.Notify([]models.CountryEntry{{Name: "Belgium", Code: "BE", Timestamp: timestamp.Add(48 * time.Hour), Confirmed: 15, Recovered: 8, Deaths: 2}})
	assert.NoError(t, err)

	err = n.Notify([]models.CountryEntry{{Name: "Belgium", Code: "BE", Timestamp: timestamp.Add(48 * time.Hour), Confirmed: 15, Recovered: 8, Deaths: 2}})
	assert.NoError(t, err)
}

func TestNotifier_Notify_Failure(t *testing.T) {
	db := mockStore.NewCovidStore(t)
	r := mockRouter.NewRouter(t)

	timestamp := time.Now()
	db.
		On("GetLatestForCountries").
		Return(map[string]models.CountryEntry{
			"Belgium": {
				Name:      "Belgium",
				Code:      "BE",
				Timestamp: timestamp,
				Confirmed: 5,
				Recovered: 1,
				Deaths:    0,
			},
		}, nil).
		Once()

	n, err := notifier.NewNotifier(r, []string{"Belgium"}, db)
	require.NoError(t, err)

	r.
		On("Send", "New probe data for Belgium", "Confirmed: 5, deaths: 1, recovered: 4").
		Return(fmt.Errorf("could not send notification")).
		Once()

	err = n.Notify([]models.CountryEntry{
		{
			Name:      "Belgium",
			Code:      "BE",
			Timestamp: timestamp.Add(24 * time.Hour),
			Confirmed: 10,
			Recovered: 5,
			Deaths:    1,
		},
	})
	assert.Error(t, err)
}

func TestNotifier_DB_Failure(t *testing.T) {
	db := mockStore.NewCovidStore(t)
	r := mockRouter.NewRouter(t)

	db.
		On("GetLatestForCountries").
		Return(nil, fmt.Errorf("db unavailable")).
		Once()

	_, err := notifier.NewNotifier(r, []string{"Belgium"}, db)
	assert.Error(t, err)
}
