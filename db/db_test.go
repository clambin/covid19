package db_test

import (
	"github.com/clambin/covid19/configuration"
	"github.com/clambin/covid19/db"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestDB_Failure(t *testing.T) {
	_, err := db.New("127.0.0.1", 5432, "test", "test", "test", nil)
	assert.Error(t, err)
}

func TestDB(t *testing.T) {
	cfg := configuration.LoadPGEnvironment()

	if !cfg.IsValid() {
		t.Log("postgres environment variables not set. skipping test")
		return
	}

	r := prometheus.NewRegistry()
	store, err := db.New(cfg.Host, cfg.Port, cfg.Database, cfg.User, cfg.Password, r)
	require.NoError(t, err)
	err = store.Handle.Ping()
	assert.NoError(t, err)
}
