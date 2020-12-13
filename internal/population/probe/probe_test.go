package probe_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	mockAPI "covid19/internal/population/apiclient/mock"
	mockDB "covid19/internal/population/db/mock"
	"covid19/internal/population/probe"
)

func TestPopulationProbe(t *testing.T) {
	db := mockDB.Create(map[string]int64{})
	apiClient := mockAPI.New(map[string]int64{
		"BE": int64(11248330),
		"US": int64(321645000),
	})

	p := probe.Create(apiClient, db)

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
