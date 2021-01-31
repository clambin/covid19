package configuration_test

import (
	"covid19/internal/configuration"
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
  rapidAPIKey: "some-key"
  notifications:
    enabled: true
    url: https://example.com/123
    countries:
      - Belgium
      - US
grafana:
  enabled: true
`

	f, err := ioutil.TempFile("", "tmp")
	if err != nil {
		panic(err)
	}

	defer os.Remove(f.Name())
	_, _ = f.Write([]byte(configString))
	_ = f.Close()

	cfg, err := configuration.LoadConfigurationFile(f.Name())

	assert.Nil(t, err)
	assert.Equal(t, 9090, cfg.Port)
	assert.True(t, cfg.Debug)
	assert.Equal(t, "localhost", cfg.Postgres.Host)
	assert.Equal(t, 31000, cfg.Postgres.Port)
	assert.Equal(t, "test", cfg.Postgres.Database)
	assert.Equal(t, "test19", cfg.Postgres.User)
	assert.Equal(t, "some-password", cfg.Postgres.Password)
	assert.True(t, cfg.Monitor.Enabled)
	assert.Equal(t, 1*time.Hour, cfg.Monitor.Interval)
	assert.Equal(t, "some-key", cfg.Monitor.RapidAPIKey)
	assert.True(t, cfg.Monitor.Notifications.Enabled)
	assert.Equal(t, "https://example.com/123", cfg.Monitor.Notifications.URL)
	if assert.Len(t, cfg.Monitor.Notifications.Countries, 2) {
		assert.Equal(t, "Belgium", cfg.Monitor.Notifications.Countries[0])
		assert.Equal(t, "US", cfg.Monitor.Notifications.Countries[1])
	}
	assert.True(t, cfg.Grafana.Enabled)
}

func TestLoadConfiguration_Defaults(t *testing.T) {
	cfg, err := configuration.LoadConfiguration([]byte{})

	assert.Nil(t, err)
	assert.Equal(t, 8080, cfg.Port)
	assert.False(t, cfg.Debug)
	assert.True(t, cfg.Monitor.Enabled)
	assert.Equal(t, 20*time.Minute, cfg.Monitor.Interval)
	assert.False(t, cfg.Grafana.Enabled)
}
