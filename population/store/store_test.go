package store_test

import (
	"github.com/clambin/covid19/configuration"
	"github.com/clambin/covid19/db"
	"github.com/clambin/covid19/population/store"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDB(t *testing.T) {
	pg := configuration.LoadPGEnvironment()
	if pg.IsValid() == false {
		return
	}

	DB, err := db.NewWithConfiguration(pg)
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
