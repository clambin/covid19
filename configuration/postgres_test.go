package configuration_test

import (
	"github.com/clambin/covid19/configuration"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func TestLoadPGEnvironment(t *testing.T) {
	cfg := configuration.LoadPGEnvironmentWithDefaults()
	assert.True(t, cfg.IsValid())
	assert.Equal(t, configuration.PostgresDB{
		Host:     "postgres",
		Port:     5432,
		Database: "covid19",
		User:     "covid",
		Password: "covid",
	}, cfg)

	cfg = configuration.LoadPGEnvironment()
	assert.False(t, cfg.IsValid())

	err := os.Setenv("pg_host", "localhost")
	require.NoError(t, err)
	err = os.Setenv("pg_port", "1234")
	require.NoError(t, err)
	err = os.Setenv("pg_database", "foo")
	require.NoError(t, err)
	err = os.Setenv("pg_user", "bar")
	require.NoError(t, err)
	err = os.Setenv("pg_password", "snafu")

	cfg = configuration.LoadPGEnvironment()
	assert.True(t, cfg.IsValid())
	assert.Equal(t, configuration.PostgresDB{
		Host:     "localhost",
		Port:     1234,
		Database: "foo",
		User:     "bar",
		Password: "snafu",
	}, cfg)
}
