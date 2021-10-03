package application_test

import (
	"context"
	"fmt"
	"github.com/clambin/covid19/application"
	"github.com/clambin/covid19/configuration"
	mockCovidFetcher "github.com/clambin/covid19/covid/probe/fetcher/mocks"
	mockCovidSaver "github.com/clambin/covid19/covid/probe/saver/mocks"
	mockCovidStore "github.com/clambin/covid19/covid/store/mocks"
	"github.com/clambin/covid19/models"
	mockPopulationProbe "github.com/clambin/covid19/population/probe/mocks"
	mockPopulationStore "github.com/clambin/covid19/population/store/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"net/http"
	"os"
	"strconv"
	"testing"
	"time"
)

func TestStack_Run_Stubbed(t *testing.T) {
	_, ok := getDBEnv()
	if ok == true {
		t.Log("Found DB env variables. Skipping this test")
		return
	}

	covidDB := &mockCovidStore.CovidStore{}
	popDB := &mockPopulationStore.PopulationStore{}

	cfg := &configuration.Configuration{
		Port:     8080,
		Debug:    false,
		Postgres: configuration.PostgresDB{},
		Monitor: configuration.MonitorConfiguration{
			Interval: 10 * time.Millisecond,
		},
	}
	stack := application.NewWithDatabases(cfg, covidDB, popDB)

	f := &mockCovidFetcher.Fetcher{}
	s := &mockCovidSaver.Saver{}
	a := &mockPopulationProbe.APIClient{}

	stack.CovidProbe.Fetcher = f
	stack.CovidProbe.Saver = s
	stack.PopulationProbe.APIClient = a

	f.On("GetCountryStats", mock.AnythingOfType("*context.cancelCtx")).Return([]*models.CountryEntry{}, nil)
	s.On("SaveNewEntries", []*models.CountryEntry{}).Return([]*models.CountryEntry{}, nil)
	a.On("GetPopulation", mock.AnythingOfType("*context.cancelCtx"), mock.AnythingOfType("string")).Return(int64(0), fmt.Errorf("API not available"))

	ctx, cancel := context.WithCancel(context.Background())
	go stack.Run(ctx)

	assert.Eventually(t, func() bool {
		resp, err := http.Get("http://localhost:8080/metrics")
		return err == nil && resp.StatusCode == http.StatusOK
	}, 500*time.Minute, 10*time.Millisecond)

	time.Sleep(200 * time.Millisecond)
	cancel()

	assert.Eventually(t, func() bool {
		resp, err := http.Get("http://localhost:8080/metrics")
		return err != nil || resp.StatusCode == http.StatusOK
	}, 500*time.Minute, 10*time.Millisecond)

	mock.AssertExpectationsForObjects(t, s, f)
}

func getDBEnv() (map[string]string, bool) {
	values := make(map[string]string, 0)
	envVars := []string{"pg_host", "pg_port", "pg_database", "pg_user", "pg_password"}

	ok := true
	for _, envVar := range envVars {
		value, found := os.LookupEnv(envVar)
		if found {
			values[envVar] = value
		} else {
			ok = false
			break
		}
	}

	return values, ok
}
func TestStack_Run(t *testing.T) {
	env, ok := getDBEnv()
	if ok == false {
		t.Log("Could not find all DB env variables. Skipping this test")
		return
	}

	port, err := strconv.Atoi(env["pg_port"])
	require.NoError(t, err)

	cfg := &configuration.Configuration{
		Port:  8080,
		Debug: false,
		Postgres: configuration.PostgresDB{
			Host:     env["pg_host"],
			Port:     port,
			Database: env["pg_database"],
			User:     env["pg_user"],
			Password: env["pg_password"],
		},
		Monitor: configuration.MonitorConfiguration{
			Interval: 10 * time.Millisecond,
		},
	}
	stack := application.New(cfg)

	f := &mockCovidFetcher.Fetcher{}
	a := &mockPopulationProbe.APIClient{}

	stack.CovidProbe.Fetcher = f
	stack.PopulationProbe.APIClient = a

	f.On("GetCountryStats", mock.AnythingOfType("*context.cancelCtx")).Return([]*models.CountryEntry{}, nil)
	a.On("GetPopulation", mock.AnythingOfType("*context.cancelCtx"), mock.AnythingOfType("string")).Return(int64(0), fmt.Errorf("population API unavailable"))

	ctx, cancel := context.WithCancel(context.Background())
	go stack.Run(ctx)

	assert.Eventually(t, func() bool {
		var resp *http.Response
		resp, err = http.Get("http://localhost:8080/metrics")
		return err == nil && resp.StatusCode == http.StatusOK
	}, 500*time.Minute, 10*time.Millisecond)

	time.Sleep(200 * time.Millisecond)
	cancel()

	assert.Eventually(t, func() bool {
		var resp *http.Response
		resp, err = http.Get("http://localhost:8080/metrics")
		return err != nil || resp.StatusCode == http.StatusOK
	}, 500*time.Minute, 10*time.Millisecond)

	mock.AssertExpectationsForObjects(t, f, a)
}
