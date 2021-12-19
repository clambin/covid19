package db_test

import (
	"github.com/clambin/covid19/db"
	"github.com/stretchr/testify/assert"
	"os"
	"strconv"
	"testing"
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

func TestDB_Stub(t *testing.T) {
	store, err := db.New("127.0.0.1", 5432, "test", "test", "test")
	assert.NoError(t, err)

	err = store.Handle.Ping()
	assert.Error(t, err)
}

func TestDB(t *testing.T) {
	values, found := getDBEnv()

	if !found {
		t.Log("postgres environment variables not set. skipping test")
		return
	}

	port, err := strconv.Atoi(values["pg_port"])
	if assert.NoError(t, err) {
		var store *db.DB
		store, err = db.New(values["pg_host"], port, values["pg_database"], values["pg_user"], values["pg_password"])
		if assert.NoError(t, err) {
			err = store.Handle.Ping()
			assert.NoError(t, err)
		}
	}
}
