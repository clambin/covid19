package main

import (
	"github.com/clambin/covid19/configuration"
	mockCovidStore "github.com/clambin/covid19/db/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"net/http"
	"testing"
	"time"
)

func TestCovidHandler(t *testing.T) {
	var stack *Stack

	pg := configuration.LoadPGEnvironment()

	if pg.IsValid() {
		cfg := &configuration.Configuration{
			Port:     8080,
			Debug:    false,
			Postgres: pg,
		}
		stack, _ = CreateStack(cfg)
	} else {
		cfg := &configuration.Configuration{
			Port:  8080,
			Debug: false,
		}
		covidStore := &mockCovidStore.CovidStore{}
		populationStore := &mockCovidStore.PopulationStore{}
		stack = CreateStackWithStores(cfg, covidStore, populationStore)
	}

	stack.SkipBackfill = true
	go stack.Run()

	assert.Eventually(t, func() bool {
		resp, err := http.Get("http://localhost:8080/metrics")
		return err == nil && resp.StatusCode == http.StatusOK
	}, 500*time.Millisecond, 10*time.Millisecond)

	stack.Stop()

	assert.Eventually(t, func() bool {
		resp, err := http.Get("http://localhost:8080/metrics")
		return err != nil || resp.StatusCode != http.StatusOK
	}, 500*time.Millisecond, 10*time.Millisecond)
}

func TestBackfill(t *testing.T) {
	cfg := &configuration.Configuration{
		Port:  8081,
		Debug: false,
	}
	covidStore := &mockCovidStore.CovidStore{}
	populationStore := &mockCovidStore.PopulationStore{}
	stack := CreateStackWithStores(cfg, covidStore, populationStore)

	covidStore.On("GetFirstEntry").Return(time.Time{}, false, nil)
	covidStore.On("Add", mock.AnythingOfType("[]models.CountryEntry")).Return(nil)

	go stack.Run()

	assert.Eventually(t, func() bool {
		resp, err := http.Get("http://localhost:8081/metrics")
		return err == nil && resp.StatusCode == http.StatusOK
	}, 500*time.Millisecond, 10*time.Millisecond)

	time.Sleep(time.Second)
}
