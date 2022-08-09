package configuration_test

import (
	"bou.ke/monkey"
	"github.com/clambin/covid19/configuration"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
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

	f, err = os.CreateTemp("", "tmp")
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
	_ = os.Setenv("RAPID_API_KEY", "")
	_ = os.Setenv("NOTIFICATION_URL", "")

	cfg, err := configuration.LoadConfiguration([]byte(configString))

	require.NoError(t, err)
	assert.Empty(t, cfg.Monitor.RapidAPIKey.Value)
	assert.Empty(t, cfg.Monitor.Notifications.URL.Value)

	_ = os.Setenv("RAPID_API_KEY", "1234")
	_ = os.Setenv("NOTIFICATION_URL", "https://example.com/")

	cfg, err = configuration.LoadConfiguration([]byte(configString))

	require.NoError(t, err)
	assert.Equal(t, "1234", cfg.Monitor.RapidAPIKey.Value)
	assert.Equal(t, "https://example.com/", cfg.Monitor.Notifications.URL.Value)
}

func TestLoadConfiguration_Defaults(t *testing.T) {
	for _, envVar := range []string{"pg_host", "pg_port", "pg_database", "pg_user", "pg_password"} {
		_ = os.Setenv(envVar, "")
	}

	cfg, err := configuration.LoadConfiguration([]byte{})

	require.NoError(t, err)
	assert.Equal(t, 8080, cfg.Port)
	assert.False(t, cfg.Debug)
	assert.Equal(t, "postgres", cfg.Postgres.Host)
	assert.Equal(t, 5432, cfg.Postgres.Port)
	assert.Equal(t, "covid19", cfg.Postgres.Database)
	assert.Equal(t, "covid", cfg.Postgres.User)
	assert.Equal(t, "covid", cfg.Postgres.Password)
}

func TestGetConfiguration(t *testing.T) {
	f, err := os.CreateTemp("", "")
	require.NoError(t, err)
	fName := f.Name()

	_, err = f.WriteString(`
port: 5000
debug: false
postgres:
  host: localhost
  port: 5555
  database: "covid19"
  user: "covid"
  password: "some-password"
monitor:
  enabled: true
  rapidAPIKey:
    value: "some-token"
  notifications:
    enabled: true
    url:
      value: "some-url"
    countries:
      - Belgium
`)
	require.NoError(t, err)
	err = f.Close()
	require.NoError(t, err)

	args := []string{"foo", "--debug", "--config", fName}
	cfg := configuration.GetConfiguration("covid19", args)
	require.NotNil(t, cfg)

	assert.Equal(t, 5000, cfg.Port)
	assert.True(t, cfg.Debug)
	assert.Equal(t, "localhost", cfg.Postgres.Host)
	assert.Equal(t, 5555, cfg.Postgres.Port)
	assert.Equal(t, "covid19", cfg.Postgres.Database)
	assert.Equal(t, "covid", cfg.Postgres.User)
	assert.Equal(t, "some-password", cfg.Postgres.Password)
	assert.Equal(t, "some-token", cfg.Monitor.RapidAPIKey.Get())
	assert.True(t, cfg.Monitor.Notifications.Enabled)
	assert.Equal(t, "some-url", cfg.Monitor.Notifications.URL.Get())
	assert.Equal(t, []string{"Belgium"}, cfg.Monitor.Notifications.Countries)

	err = os.Remove(fName)
	require.NoError(t, err)
}

func TestGetConfiguration_Invalid(t *testing.T) {
	var exited bool
	monkey.Patch(os.Exit, func(int) { exited = true })
	args := []string{"foo", "--?"}
	_ = configuration.GetConfiguration("covid19", args)
	assert.True(t, exited)
	monkey.Unpatch(os.Exit)
}
