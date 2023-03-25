package covid_test

import (
	"github.com/clambin/covid19/covid"
	"github.com/clambin/covid19/covid/shoutrrr/mocks"
	"github.com/clambin/covid19/models"
	"github.com/clambin/go-common/set"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestNotifier_Notify(t *testing.T) {
	s := mocks.NewSender(t)
	c := covid.Notifier{
		Countries: set.Create("Belgium", "US"),
		Sender:    s,
	}

	current := map[string]models.CountryEntry{
		"Belgium": {
			Timestamp: time.Date(2023, time.March, 21, 0, 0, 0, 0, time.UTC),
			Code:      "BE",
			Name:      "Belgium",
			Confirmed: 100,
			Recovered: 50,
			Deaths:    25,
		},
	}
	update := []models.CountryEntry{
		{
			Timestamp: time.Date(2023, time.March, 22, 0, 0, 0, 0, time.UTC),
			Code:      "BE",
			Name:      "Belgium",
			Confirmed: 200,
			Recovered: 100,
			Deaths:    50,
		},
		{
			Timestamp: time.Date(2023, time.March, 22, 0, 0, 0, 0, time.UTC),
			Code:      "FR",
			Name:      "France",
			Confirmed: 2000,
			Recovered: 1000,
			Deaths:    500,
		},
	}

	s.On("Send", "New data for Belgium", "Confirmed: 100, deaths: 25").Return(nil)
	err := c.Notify(current, update)
	require.NoError(t, err)
}
