package covidprobe_test

import (
	"context"
	"github.com/clambin/covid19/configuration"
	"github.com/clambin/covid19/coviddb"
	dbMock "github.com/clambin/covid19/coviddb/mocks"
	"github.com/clambin/covid19/covidprobe"
	probeMock "github.com/clambin/covid19/covidprobe/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

var (
	lastUpdate = time.Date(2020, time.December, 3, 5, 28, 22, 0, time.UTC)
)

func TestProbe_Run(t *testing.T) {
	cfg := &configuration.MonitorConfiguration{
		Interval:      25 * time.Millisecond,
		RapidAPIKey:   configuration.ValueOrEnvVar{Value: "akey"},
		Notifications: configuration.NotificationConfiguration{},
	}
	db := &dbMock.DB{}
	apiClient := &probeMock.APIClient{}

	probe := covidprobe.NewProbe(cfg, db, nil)
	probe.APIClient = apiClient
	probe.TestMode = true

	timestamp := time.Now()

	// setup expectations
	apiClient.On("GetCountryStats", mock.Anything).Return(map[string]covidprobe.CountryStats{
		"Belgium":     {LastUpdate: timestamp, Confirmed: 40, Recovered: 10, Deaths: 1},
		"US":          {LastUpdate: timestamp, Confirmed: 20, Recovered: 15, Deaths: 5},
		"NotACountry": {LastUpdate: timestamp, Confirmed: 0, Recovered: 0, Deaths: 0},
	}, nil)
	db.On("ListLatestByCountry").Return(map[string]time.Time{}, nil).Once()
	db.On("ListLatestByCountry").Return(map[string]time.Time{"Belgium": timestamp, "US": timestamp}, nil)
	db.On("Add", []coviddb.CountryEntry{
		{Code: "BE", Name: "Belgium", Timestamp: timestamp, Confirmed: 40, Recovered: 10, Deaths: 1},
		{Code: "US", Name: "US", Timestamp: timestamp, Confirmed: 20, Recovered: 15, Deaths: 5},
	}).Return(nil).Once()

	// log.SetLevel(log.DebugLevel)

	// Run the probe
	err := probe.Update(context.Background())
	require.NoError(t, err)
	mock.AssertExpectationsForObjects(t, apiClient, db)
}
