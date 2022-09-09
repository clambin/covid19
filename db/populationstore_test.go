package db_test

import (
	"github.com/clambin/covid19/configuration"
	"github.com/clambin/covid19/db"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestPopulationStore(t *testing.T) {
	pg := configuration.LoadPGEnvironment()
	if pg.IsValid() == false {
		return
	}

	r := prometheus.NewRegistry()
	db2, err := db.NewWithConfiguration(pg, r)
	require.NoError(t, err)

	popDB := db.NewPopulationStore(db2)

	_, err = popDB.List()
	assert.NoError(t, err)

	err = popDB.Add("???", 242)
	require.NoError(t, err)

	newContent, err := popDB.List()
	assert.NoError(t, err)

	entry, ok := newContent["???"]
	assert.True(t, ok)
	assert.Equal(t, int64(242), entry)
}
