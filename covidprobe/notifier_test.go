package covidprobe_test

import (
	"bou.ke/monkey"

	"fmt"
	"github.com/clambin/covid19/coviddb"
	dbMock "github.com/clambin/covid19/coviddb/mocks"
	"github.com/clambin/covid19/covidprobe"
	probeMock "github.com/clambin/covid19/covidprobe/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"os"
	"testing"
	"time"
)

func TestNotifier_Notify(t *testing.T) {
	db := &dbMock.DB{}
	router := &probeMock.NotificationSender{}

	timestamp := time.Now()
	db.On("GetLastForCountry", "Belgium").Return(&coviddb.CountryEntry{
		Name:      "Belgium",
		Code:      "BE",
		Timestamp: timestamp,
		Confirmed: 5,
		Recovered: 1,
		Deaths:    0,
	}, true, nil)

	notifier := covidprobe.NewNotifier(router, []string{"Belgium"}, db)

	// empty db.  no notifications to be sent
	err := notifier.Notify([]coviddb.CountryEntry{})
	assert.NoError(t, err)

	router.On("Send", "New covid data for Belgium", "Confirmed: 5, deaths: 1, recovered: 4").Return(nil).Once()
	err = notifier.Notify([]coviddb.CountryEntry{
		{
			Name:      "Belgium",
			Code:      "BE",
			Timestamp: timestamp.Add(24 * time.Hour),
			Confirmed: 10,
			Recovered: 5,
			Deaths:    1,
		},
	})
	assert.NoError(t, err)

	router.On("Send", "New covid data for Belgium", "Confirmed: 5, deaths: 1, recovered: 3").Return(nil).Once()
	err = notifier.Notify([]coviddb.CountryEntry{
		{
			Name:      "Belgium",
			Code:      "BE",
			Timestamp: timestamp.Add(48 * time.Hour),
			Confirmed: 15,
			Recovered: 8,
			Deaths:    2,
		},
	})
	assert.NoError(t, err)

	mock.AssertExpectationsForObjects(t, db, router)
}

func TestNotifier_Notify_Failure(t *testing.T) {
	db := &dbMock.DB{}
	router := &probeMock.NotificationSender{}

	timestamp := time.Now()
	db.On("GetLastForCountry", "Belgium").Return(&coviddb.CountryEntry{
		Name:      "Belgium",
		Code:      "BE",
		Timestamp: timestamp,
		Confirmed: 5,
		Recovered: 1,
		Deaths:    0,
	}, true, nil)

	notifier := covidprobe.NewNotifier(router, []string{"Belgium"}, db)

	router.On("Send", "New covid data for Belgium", "Confirmed: 5, deaths: 1, recovered: 4").Return(fmt.Errorf("could not send notification"))
	err := notifier.Notify([]coviddb.CountryEntry{
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
	mock.AssertExpectationsForObjects(t, db, router)
}

func TestNotifier_DB_Failure(t *testing.T) {
	db := &dbMock.DB{}
	router := &probeMock.NotificationSender{}

	db.On("GetLastForCountry", "Belgium").Return(&coviddb.CountryEntry{}, false, fmt.Errorf("db unavailable"))

	fakeExit := func(int) {
		panic("os.Exit called")
	}
	patch := monkey.Patch(os.Exit, fakeExit)
	defer patch.Unpatch()

	assert.Panics(t, func() {
		_ = covidprobe.NewNotifier(router, []string{"Belgium"}, db)
	})

	mock.AssertExpectationsForObjects(t, db, router)
}
