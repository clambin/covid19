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
	assert.Equal(t, 9090, cfg.Port)
	assert.True(t, cfg.Debug)
	assert.Equal(t, "localhost", cfg.Postgres.Host)
	assert.Equal(t, 31000, cfg.Postgres.Port)
	assert.Equal(t, "test", cfg.Postgres.Database)
	assert.Equal(t, "test19", cfg.Postgres.User)
	assert.Equal(t, "some-password", cfg.Postgres.Password)
	assert.Equal(t, "some-key", cfg.Monitor.RapidAPIKey.Get())
	assert.True(t, cfg.Monitor.Notifications.Enabled)
	assert.Equal(t, "https://example.com/123", cfg.Monitor.Notifications.URL.Get())
	require.Len(t, cfg.Monitor.Notifications.Countries, 2)
	assert.Equal(t, "Belgium", cfg.Monitor.Notifications.Countries[0])
	assert.Equal(t, "US", cfg.Monitor.Notifications.Countries[1])
}

func TestLoadConfiguration_EnvVars(t *testing.T) {
	const configString = `
monitor:
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
	_ = os.Setenv("RAPID_API_KEY", "")
	_ = os.Setenv("NOTIFICATION_URL", "")

	cfg, err := configuration.LoadConfiguration(bytes.NewBufferString(configString))

	require.NoError(t, err)
	assert.Empty(t, cfg.Monitor.RapidAPIKey.Value)
	assert.Empty(t, cfg.Monitor.Notifications.URL.Value)

	_ = os.Setenv("RAPID_API_KEY", "1234")
	_ = os.Setenv("NOTIFICATION_URL", "https://example.com/")

	cfg, err = configuration.LoadConfiguration(bytes.NewBufferString(configString))

	require.NoError(t, err)
	assert.Equal(t, "1234", cfg.Monitor.RapidAPIKey.Value)
	assert.Equal(t, "https://example.com/", cfg.Monitor.Notifications.URL.Value)
}

func TestLoadConfiguration_Defaults(t *testing.T) {
	for _, envVar := range []string{"pg_host", "pg_port", "pg_database", "pg_user", "pg_password"} {
		_ = os.Setenv(envVar, "")
	}

	content := bytes.NewBufferString("foo: bar")
	cfg, err := configuration.LoadConfiguration(content)

	require.NoError(t, err)
	assert.Equal(t, 8080, cfg.Port)
	assert.False(t, cfg.Debug)
	assert.Equal(t, "postgres", cfg.Postgres.Host)
	assert.Equal(t, 5432, cfg.Postgres.Port)
	assert.Equal(t, "covid19", cfg.Postgres.Database)
	assert.Equal(t, "covid", cfg.Postgres.User)
	assert.Equal(t, "covid", cfg.Postgres.Password)
}
