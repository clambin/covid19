package configuration_test

import (
	"github.com/clambin/covid19/configuration"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"testing"
	"time"
)

func TestLoadConfigurationFile(t *testing.T) {
	const configString = `
port: 9090
debug: true
postgres:
  host: localhost
  port: 31000
  database: "test"
  user: "test19"
  password: "some-password"
monitor:
  enabled: true
  interval: 1h
  rapidAPIKey: 
    value: "some-key"
  notifications:
    enabled: true
    url: 
      value: https://example.com/123
    countries:
      - Belgium
      - US
grafana:
  enabled: true
`
	var (
		err error
		f   *os.File
		cfg *configuration.Configuration
	)

	f, err = ioutil.TempFile("", "tmp")
	if err != nil {
		panic(err)
	}

	defer func(filename string) {
		_ = os.Remove(filename)
	}(f.Name())
	_, _ = f.Write([]byte(configString))
	_ = f.Close()

	cfg, err = configuration.LoadConfigurationFile(f.Name())

	if assert.Nil(t, err) {
		assert.Equal(t, 9090, cfg.Port)
		assert.True(t, cfg.Debug)
		assert.Equal(t, "localhost", cfg.Postgres.Host)
		assert.Equal(t, 31000, cfg.Postgres.Port)
		assert.Equal(t, "test", cfg.Postgres.Database)
		assert.Equal(t, "test19", cfg.Postgres.User)
		assert.Equal(t, "some-password", cfg.Postgres.Password)
		assert.Equal(t, 1*time.Hour, cfg.Monitor.Interval)
		assert.Equal(t, "some-key", cfg.Monitor.RapidAPIKey.Value)
		assert.True(t, cfg.Monitor.Notifications.Enabled)
		assert.Equal(t, "https://example.com/123", cfg.Monitor.Notifications.URL.Value)
		if assert.Len(t, cfg.Monitor.Notifications.Countries, 2) {
			assert.Equal(t, "Belgium", cfg.Monitor.Notifications.Countries[0])
			assert.Equal(t, "US", cfg.Monitor.Notifications.Countries[1])
		}
	}
}

func TestLoadConfiguration_EnvVars(t *testing.T) {
	const configString = `
monitor:
  enabled: true
  rapidAPIKey: 
    envVar: "RAPID_API_KEY"
  notifications:
    enabled: true
    url: 
      envVar: "NOTIFICATION_URL"
    countries:
      - Belgium
      - US
`
	var (
		err error
		cfg *configuration.Configuration
	)

	_ = os.Setenv("RAPID_API_KEY", "")
	_ = os.Setenv("NOTIFICATION_URL", "")

	cfg, err = configuration.LoadConfiguration([]byte(configString))

	if assert.Nil(t, err) {
		assert.Empty(t, cfg.Monitor.RapidAPIKey.Value)
		assert.Empty(t, cfg.Monitor.Notifications.URL.Value)
	}

	_ = os.Setenv("RAPID_API_KEY", "1234")
	_ = os.Setenv("NOTIFICATION_URL", "https://example.com/")

	cfg, err = configuration.LoadConfiguration([]byte(configString))

	if assert.Nil(t, err) {
		assert.Equal(t, "1234", cfg.Monitor.RapidAPIKey.Value)
		assert.Equal(t, "https://example.com/", cfg.Monitor.Notifications.URL.Value)
	}
}

func TestLoadConfiguration_Defaults(t *testing.T) {
	for _, envVar := range []string{"pg_host", "pg_port", "pg_database", "pg_user", "pg_password"} {
		_ = os.Setenv(envVar, "")
	}

	cfg, err := configuration.LoadConfiguration([]byte{})

	assert.Nil(t, err)
	assert.Equal(t, 8080, cfg.Port)
	assert.False(t, cfg.Debug)
	assert.Equal(t, 20*time.Minute, cfg.Monitor.Interval)
	assert.Equal(t, "postgres", cfg.Postgres.Host)
	assert.Equal(t, 5432, cfg.Postgres.Port)
	assert.Equal(t, "covid19", cfg.Postgres.Database)
	assert.Equal(t, "probe", cfg.Postgres.User)
	assert.Equal(t, "probe", cfg.Postgres.Password)
}
