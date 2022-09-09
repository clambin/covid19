package fetcher_test

import (
	"github.com/clambin/covid19/covid/fetcher"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestStringSet(t *testing.T) {
	ss := fetcher.StringSet{}

	found := ss.IsSet("foo")
	assert.False(t, found)

	found = ss.Set("foo")
	assert.False(t, found)

	found = ss.IsSet("foo")
	assert.True(t, found)

	found = ss.Set("foo")
	assert.True(t, found)
}
