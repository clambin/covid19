package configuration_test

import (
	"bytes"
	"github.com/clambin/covid19/configuration"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
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

	body, err := yaml.Marshal(&cfg)
	require.NoError(t, err)
	assert.Equal(t, `port: 9090
prometheusPort: 9092
debug: true
postgres:
    host: localhost
    port: 31000
    database: test
    user: test19
    password: some-password
monitor:
    rapidAPIKey: some-key
    notifications:
        enabled: true
        url: https://example.com/123
        countries:
            - Belgium
            - US
`, string(body))
}

func TestLoadConfiguration_Defaults(t *testing.T) {
	for _, envVar := range []string{"pg_host", "pg_port", "pg_database", "pg_user", "pg_password"} {
		_ = os.Setenv(envVar, "")
	}

	content := bytes.NewBufferString("foo: bar")
	cfg, err := configuration.LoadConfiguration(content)
	require.NoError(t, err)

	body, err := yaml.Marshal(&cfg)
	require.NoError(t, err)
	assert.Equal(t, `port: 8080
prometheusPort: 9090
debug: false
postgres:
    host: postgres
    port: 5432
    database: covid19
    user: covid
    password: ""
monitor:
    rapidAPIKey: ""
    notifications:
        enabled: false
        url: ""
        countries: []
`, string(body))
}
