package configuration_test

import (
	"bytes"
	"github.com/clambin/covid19/configuration"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func TestLoadConfiguration(t *testing.T) {
	const configString = `
port: 9090
prometheusPort: 9092
debug: true
postgres:
  host: localhost
  port: 31000
  database: "test"
  user: "test19"
  password: "$pg_password"
monitor:
  interval: 1h
  rapidAPIKey: "some-key"
  notifications:
    enabled: true
    url: https://example.com/123
    countries:
      - Belgium
      - US
`

	err := os.Setenv("pg_password", "some-password")
	require.NoError(t, err)

	cfg, err := configuration.LoadConfiguration(bytes.NewBufferString(configString))
	require.NoError(t, err)

	assert.True(t, cfg.Postgres.IsValid())

	assert.Equal(t, 9090, cfg.Port)
	assert.Equal(t, 9092, cfg.PrometheusPort)
	assert.True(t, cfg.Debug)
	assert.Equal(t, "localhost", cfg.Postgres.Host)
	assert.Equal(t, 31000, cfg.Postgres.Port)
	assert.Equal(t, "test", cfg.Postgres.Database)
	assert.Equal(t, "test19", cfg.Postgres.User)
	assert.Equal(t, "some-password", cfg.Postgres.Password)
	assert.Equal(t, "some-key", cfg.Monitor.RapidAPIKey)
	assert.True(t, cfg.Monitor.Notifications.Enabled)
	assert.Equal(t, "https://example.com/123", cfg.Monitor.Notifications.URL)
	require.Len(t, cfg.Monitor.Notifications.Countries, 2)
	assert.Equal(t, "Belgium", cfg.Monitor.Notifications.Countries[0])
	assert.Equal(t, "US", cfg.Monitor.Notifications.Countries[1])
}

func TestLoadConfiguration_Defaults(t *testing.T) {
	for _, envVar := range []string{"pg_host", "pg_port", "pg_database", "pg_user", "pg_password"} {
		_ = os.Setenv(envVar, "")
	}

	content := bytes.NewBufferString("foo: bar")
	cfg, err := configuration.LoadConfiguration(content)
	require.NoError(t, err)

	assert.Equal(t, 8080, cfg.Port)
	assert.Equal(t, 9090, cfg.PrometheusPort)
	assert.False(t, cfg.Debug)
	assert.Equal(t, "postgres", cfg.Postgres.Host)
	assert.Equal(t, 5432, cfg.Postgres.Port)
	assert.Equal(t, "covid19", cfg.Postgres.Database)
	assert.Equal(t, "covid", cfg.Postgres.User)
	assert.Equal(t, "", cfg.Postgres.Password)
	assert.False(t, cfg.Monitor.Notifications.Enabled)
}
