package testdb

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"covid19/internal/coviddb"
)

func TestDBCache_List(t *testing.T) {
	dbc := coviddb.NewCache(CreateWithData(), 5*time.Minute)

	assert.NotNil(t, dbc)

	entries, err := dbc.List(time.Now())
	assert.Nil(t, err)
	assert.Equal(t, 4, len(entries))

	entries, err = dbc.List(time.Date(2020, time.November, 2, 0, 0, 0, 0, time.UTC))
	assert.Nil(t, err)
	assert.Equal(t, 1, len(entries))
	assert.Equal(t, "BE", entries[0].Code)
	assert.Equal(t, int64(1), entries[0].Confirmed)
	assert.Equal(t, int64(0), entries[0].Deaths)
	assert.Equal(t, int64(0), entries[0].Recovered)
}
