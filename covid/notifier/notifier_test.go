package notifier_test

import (
	"bou.ke/monkey"
	"fmt"
	"github.com/clambin/covid19/covid/notifier"
	mockNotificationSender "github.com/clambin/covid19/covid/notifier/mocks"
	mockStore "github.com/clambin/covid19/db/mocks"
	"github.com/clambin/covid19/models"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"time"
)

func TestNotifier_Notify(t *testing.T) {
	db := mockStore.NewCovidStore(t)
	sender := mockNotificationSender.NewNotificationSender(t)

	timestamp := time.Now()
	db.
		On("GetLatestForCountries", []string{"Belgium"}).
		Return(map[string]models.CountryEntry{"Belgium": {Name: "Belgium", Code: "BE", Timestamp: timestamp, Confirmed: 5, Recovered: 1, Deaths: 0}}, nil).
		Once()

	n := notifier.NewNotifier(sender, []string{"Belgium"}, db)

	err := n.Notify([]models.CountryEntry{})
	assert.NoError(t, err)

	err = n.Notify([]models.CountryEntry{
		{Name: "Belgium", Code: "BE", Timestamp: timestamp.Add(24 * time.Hour), Confirmed: 5, Recovered: 1, Deaths: 0},
	})
	assert.NoError(t, err)

	sender.On("Send", "New probe data for Belgium", "Confirmed: 5, deaths: 1, recovered: 4").Return(nil).Once()
	err = n.Notify([]models.CountryEntry{
		{Name: "Belgium", Code: "BE", Timestamp: timestamp.Add(24 * time.Hour), Confirmed: 10, Recovered: 5, Deaths: 1},
		{Name: "US", Code: "US", Timestamp: timestamp, Confirmed: 50, Recovered: 10, Deaths: 5},
	})
	assert.NoError(t, err)

	sender.On("Send", "New probe data for Belgium", "Confirmed: 5, deaths: 1, recovered: 3").Return(nil).Once()
	err = n.Notify([]models.CountryEntry{{Name: "Belgium", Code: "BE", Timestamp: timestamp.Add(48 * time.Hour), Confirmed: 15, Recovered: 8, Deaths: 2}})
	assert.NoError(t, err)

	err = n.Notify([]models.CountryEntry{{Name: "Belgium", Code: "BE", Timestamp: timestamp.Add(48 * time.Hour), Confirmed: 15, Recovered: 8, Deaths: 2}})
	assert.NoError(t, err)
}

func TestNotifier_Notify_Failure(t *testing.T) {
	db := mockStore.NewCovidStore(t)
	router := mockNotificationSender.NewNotificationSender(t)

	timestamp := time.Now()
	db.
		On("GetLatestForCountries", []string{"Belgium"}).
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

	n := notifier.NewNotifier(router, []string{"Belgium"}, db)

	router.
		On("Send", "New probe data for Belgium", "Confirmed: 5, deaths: 1, recovered: 4").
		Return(fmt.Errorf("could not send notification")).
		Once()

	err := n.Notify([]models.CountryEntry{
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
	router := mockNotificationSender.NewNotificationSender(t)

	db.
		On("GetLatestForCountries", []string{"Belgium"}).
		Return(nil, fmt.Errorf("db unavailable")).
		Once()

	fakeExit := func(int) {
		panic("os.Exit called")
	}
	patch := monkey.Patch(os.Exit, fakeExit)
	defer patch.Unpatch()

	assert.Panics(t, func() {
		_ = notifier.NewNotifier(router, []string{"Belgium"}, db)
	})
}
