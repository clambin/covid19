package db_test

import (
	"github.com/clambin/covid19/configuration"
	"github.com/clambin/covid19/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestDB_Stub(t *testing.T) {
	store, err := db.New("127.0.0.1", 5432, "test", "test", "test")
	assert.NoError(t, err)

	err = store.Handle.Ping()
	assert.Error(t, err)
}

func TestDB(t *testing.T) {
	cfg := configuration.LoadPGEnvironment()

	if !cfg.IsValid() {
		t.Log("postgres environment variables not set. skipping test")
		return
	}

	store, err := db.New(cfg.Host, cfg.Port, cfg.Database, cfg.User, cfg.Password)
	require.NoError(t, err)
	err = store.Handle.Ping()
	assert.NoError(t, err)
}
