package store_test

import (
	"github.com/clambin/covid19/db"
	"github.com/clambin/covid19/population/store"
	"github.com/stretchr/testify/assert"
	"os"
	"strconv"
	"testing"
)

func getDBEnv() (map[string]string, bool) {
	values := make(map[string]string, 0)
	envVars := []string{"pg_host", "pg_port", "pg_database", "pg_user", "pg_password"}

	ok := true
	for _, envVar := range envVars {
		value, found := os.LookupEnv(envVar)
		if found {
			values[envVar] = value
		} else {
			ok = false
			break
		}
	}

	return values, ok
}

func TestDB(t *testing.T) {
	values, ok := getDBEnv()
	if ok == false {
		return
	}

	port, err := strconv.Atoi(values["pg_port"])
	assert.Nil(t, err)

	var DB *db.DB
	DB, err = db.New(values["pg_host"], port, values["pg_database"], values["pg_user"], values["pg_password"])
	assert.NoError(t, err)

	var popDB store.PopulationStore
	popDB = store.New(DB)

	_, err = popDB.List()
	assert.Nil(t, err)

	err = popDB.Add("???", 242)
	assert.Nil(t, err)

	newContent, err := popDB.List()
	assert.Nil(t, err)

	entry, ok := newContent["???"]
	assert.True(t, ok)
	assert.Equal(t, int64(242), entry)
}
