package db_test

import (
	"github.com/clambin/covid19/internal/population/db"
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

	covidDB := db.NewPostgresDB(values["pg_host"], port, values["pg_database"], values["pg_user"], values["pg_password"])
	assert.NotNil(t, covidDB)

	_, err = covidDB.List()
	assert.Nil(t, err)

	err = covidDB.Add(map[string]int64{"???": 242})
	assert.Nil(t, err)

	newContent, err := covidDB.List()
	assert.Nil(t, err)

	entry, ok := newContent["???"]
	assert.True(t, ok)
	assert.Equal(t, int64(242), entry)
}
