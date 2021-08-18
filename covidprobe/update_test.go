package covidprobe_test

import (
	"context"
	"fmt"
	"github.com/clambin/covid19/configuration"
	"github.com/clambin/covid19/coviddb"
	dbMock "github.com/clambin/covid19/coviddb/mocks"
	"github.com/clambin/covid19/covidprobe"
	probeMock "github.com/clambin/covid19/covidprobe/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestProbe_Update_NoData(t *testing.T) {
	cfg := &configuration.MonitorConfiguration{
		Interval:      30 * time.Minute,
		RapidAPIKey:   configuration.ValueOrEnvVar{Value: "akey"},
		Notifications: configuration.NotificationConfiguration{},
	}
	db := &dbMock.DB{}
	apiClient := &probeMock.APIClient{}

	probe := covidprobe.NewProbe(cfg, db, nil)
	probe.APIClient = apiClient
	probe.TestMode = true

	timestamp := time.Now()

	// Initial update; nothing in database. Valid countries are inserted.
	apiClient.On("GetCountryStats", mock.Anything).Return(map[string]covidprobe.CountryStats{
		"Belgium":     {LastUpdate: timestamp, Confirmed: 40, Recovered: 10, Deaths: 1},
		"US":          {LastUpdate: timestamp, Confirmed: 20, Recovered: 15, Deaths: 5},
		"NotACountry": {LastUpdate: timestamp, Confirmed: 0, Recovered: 0, Deaths: 0},
	}, nil)
	db.On("ListLatestByCountry").Return(map[string]time.Time{}, nil).Once()
	db.On("Add", []coviddb.CountryEntry{
		{Code: "BE", Name: "Belgium", Timestamp: timestamp, Confirmed: 40, Recovered: 10, Deaths: 1},
		{Code: "US", Name: "US", Timestamp: timestamp, Confirmed: 20, Recovered: 15, Deaths: 5},
	}).Return(nil).Once()

	err := probe.Update(context.TODO())
	require.NoError(t, err)
	mock.AssertExpectationsForObjects(t, apiClient, db)

	db.On("ListLatestByCountry").Return(map[string]time.Time{"Belgium": timestamp, "US": timestamp}, nil).Once()
	err = probe.Update(context.TODO())
	require.NoError(t, err)
	mock.AssertExpectationsForObjects(t, apiClient, db)
}

func TestProbe_Update_WithData(t *testing.T) {
	cfg := &configuration.MonitorConfiguration{
		Interval:      30 * time.Minute,
		RapidAPIKey:   configuration.ValueOrEnvVar{Value: "akey"},
		Notifications: configuration.NotificationConfiguration{},
	}
	db := &dbMock.DB{}
	apiClient := &probeMock.APIClient{}

	probe := covidprobe.NewProbe(cfg, db, nil)
	probe.APIClient = apiClient

	// Data already in database. No inserts are done.
	apiClient.On("GetCountryStats", mock.Anything).Return(map[string]covidprobe.CountryStats{
		"Belgium":     {LastUpdate: lastUpdate, Confirmed: 40, Recovered: 10, Deaths: 1},
		"US":          {LastUpdate: lastUpdate, Confirmed: 20, Recovered: 15, Deaths: 5},
		"NotACountry": {LastUpdate: lastUpdate, Confirmed: 0, Recovered: 0, Deaths: 0},
	}, nil)
	db.On("ListLatestByCountry").Return(map[string]time.Time{"Belgium": lastUpdate, "US": lastUpdate}, nil)

	err := probe.Update(context.TODO())
	require.NoError(t, err)
	mock.AssertExpectationsForObjects(t, apiClient, db)
}

func TestProbe_Update_WithNotifier(t *testing.T) {
	cfg := &configuration.MonitorConfiguration{
		Interval:    30 * time.Minute,
		RapidAPIKey: configuration.ValueOrEnvVar{Value: "akey"},
		Notifications: configuration.NotificationConfiguration{
			Enabled:   true,
			Countries: []string{"Belgium"},
			URL:       configuration.ValueOrEnvVar{Value: "https://example.com"},
		},
	}
	db := &dbMock.DB{}
	apiClient := &probeMock.APIClient{}
	sender := &probeMock.NotificationSender{}

	apiClient.On("GetCountryStats", mock.Anything).Return(map[string]covidprobe.CountryStats{
		"Belgium":     {LastUpdate: lastUpdate, Confirmed: 40, Recovered: 10, Deaths: 1},
		"US":          {LastUpdate: lastUpdate, Confirmed: 20, Recovered: 15, Deaths: 5},
		"NotACountry": {LastUpdate: lastUpdate, Confirmed: 0, Recovered: 0, Deaths: 0},
	}, nil)
	db.On("GetLastForCountry", "Belgium").Return(&coviddb.CountryEntry{
		Timestamp: lastUpdate.Add(-24 * time.Hour),
		Code:      "BE",
		Name:      "Belgium",
		Confirmed: 35,
		Recovered: 5,
		Deaths:    0,
	}, true, nil)
	db.On("ListLatestByCountry").Return(map[string]time.Time{"Belgium": lastUpdate.Add(-24 * time.Hour)}, nil)
	db.On("Add", []coviddb.CountryEntry{
		{Code: "BE", Name: "Belgium", Timestamp: lastUpdate, Confirmed: 40, Recovered: 10, Deaths: 1},
		{Code: "US", Name: "US", Timestamp: lastUpdate, Confirmed: 20, Recovered: 15, Deaths: 5},
	}).Return(nil).Once()
	sender.On("Send", "New covid data for Belgium", "Confirmed: 5, deaths: 1, recovered: 5").Return(nil)

	probe := covidprobe.NewProbe(cfg, db, nil)
	probe.TestMode = true
	probe.APIClient = apiClient
	probe.Notifier = covidprobe.NewNotifier(sender, []string{"Belgium"}, db)

	err := probe.Update(context.TODO())
	require.NoError(t, err)
	mock.AssertExpectationsForObjects(t, db, apiClient, sender)

}

func TestProbe_Update_APIFailure(t *testing.T) {
	cfg := &configuration.MonitorConfiguration{
		Interval:      30 * time.Minute,
		RapidAPIKey:   configuration.ValueOrEnvVar{Value: "akey"},
		Notifications: configuration.NotificationConfiguration{},
	}
	db := &dbMock.DB{}
	apiClient := &probeMock.APIClient{}

	probe := covidprobe.NewProbe(cfg, db, nil)
	probe.APIClient = apiClient

	apiClient.On("GetCountryStats", mock.Anything).Return(map[string]covidprobe.CountryStats{}, fmt.Errorf("API not available")).Once()

	err := probe.Update(context.TODO())
	assert.Error(t, err)
	assert.Equal(t, "failed to get Covid figures: API not available", err.Error())
	//apiClient.AssertExpectations(t)

	apiClient.On("GetCountryStats", mock.Anything).Return(map[string]covidprobe.CountryStats{
		"Belgium":     {LastUpdate: lastUpdate, Confirmed: 40, Recovered: 10, Deaths: 1},
		"US":          {LastUpdate: lastUpdate, Confirmed: 20, Recovered: 15, Deaths: 5},
		"NotACountry": {LastUpdate: lastUpdate, Confirmed: 0, Recovered: 0, Deaths: 0},
	}, nil)
	db.On("ListLatestByCountry").Return(map[string]time.Time{}, fmt.Errorf("database is down"))

	err = probe.Update(context.TODO())
	assert.Error(t, err)
	assert.Equal(t, "failed to process Covid figures: database is down", err.Error())

	mock.AssertExpectationsForObjects(t, apiClient, db)
}
