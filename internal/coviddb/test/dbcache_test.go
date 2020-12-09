package test

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
	assert.NotZero(t, len(entries))
}
