package db_test

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestPopulationStore(t *testing.T) {
	_, err := popStore.List()
	assert.NoError(t, err)

	err = popStore.Add("???", 242)
	require.NoError(t, err)

	newContent, err := popStore.List()
	assert.NoError(t, err)

	entry, ok := newContent["???"]
	assert.True(t, ok)
	assert.Equal(t, int64(242), entry)
}
