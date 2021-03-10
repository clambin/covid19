package probe_test

import (
	mockDB "github.com/clambin/covid19/internal/population/db/mock"
	"github.com/clambin/covid19/internal/population/probe"
	"github.com/clambin/covid19/internal/population/probe/mockapi"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPopulationProbe(t *testing.T) {
	db := mockDB.Create(map[string]int64{})

	p := probe.Create("", db)
	p.APIClient = mockapi.New(map[string]int64{
		"BE": int64(11248330),
		"US": int64(321645000),
	})

	// DB should be empty
	entries, err := db.List()
	assert.Nil(t, err)
	assert.Len(t, entries, 0)

	// Run the probe
	err = p.Run()
	assert.Nil(t, err)

	// Check results
	entries, err = db.List()
	assert.Nil(t, err)
	assert.Len(t, entries, 2)

	val, ok := entries["BE"]
	assert.True(t, ok)
	assert.Equal(t, int64(11248330), val)

	val, ok = entries["US"]
	assert.True(t, ok)
	assert.Equal(t, int64(321645000), val)
}
