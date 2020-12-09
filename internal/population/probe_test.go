package population

import (
	"github.com/stretchr/testify/assert"
	"testing"

	"covid19/internal/population/mock"
)

func TestPopulationProbe(t *testing.T) {
	db := mock.Create(testDBData)
	probe := Create(makeClient(), db)

	// Basic mock db testtools
	entries, err := db.List()
	assert.Nil(t, err)
	assert.Len(t, entries, 1)

	val, ok := entries["BE"]
	assert.True(t, ok)
	assert.Equal(t, int64(1), val)

	// Run the probe
	assert.Nil(t, probe.Run())

	// Check results
	entries, err = db.List()
	assert.Nil(t, err)
	assert.Len(t, entries, 2)

	val, ok = entries["BE"]
	assert.True(t, ok)
	assert.Equal(t, int64(11248330), val)

	val, ok = entries["US"]
	assert.True(t, ok)
	assert.Equal(t, int64(321645000), val)
}
