package main

import (
	"github.com/clambin/covid19/configuration"
	"github.com/clambin/covid19/covid/store/mocks"
	"github.com/stretchr/testify/assert"
	"net/http"
	"os"
	"strconv"
	"testing"
	"time"
)

func getDBEnv() (map[string]string, bool) {
	values := make(map[string]string, 0)
	envVars := []string{"pg_host", "pg_port", "pg_database", "pg_user", "pg_password"}

	for _, envVar := range envVars {
		value, found := os.LookupEnv(envVar)
		if !found {
			return nil, false
		}
		values[envVar] = value
	}

	return values, true
}

func TestCovidHandler(t *testing.T) {
	var stack *Stack

	vars, ok := getDBEnv()

	if ok {
		port, _ := strconv.Atoi(vars["pg_port"])

		cfg := &configuration.Configuration{
			Port:  8080,
			Debug: false,
			Postgres: configuration.PostgresDB{
				Host:     vars["pg_host"],
				Port:     port,
				Database: vars["pg_database"],
				User:     vars["pg_user"],
				Password: vars["pg_password"],
			},
		}
		stack = CreateStack(cfg)
	} else {
		cfg := &configuration.Configuration{
			Port:  8080,
			Debug: false,
		}
		store := &mocks.CovidStore{}
		stack = CreateStackWithStore(cfg, store)
	}

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