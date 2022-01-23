package store

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestTimestampClause(t *testing.T) {
	testCases := []struct {
		from     time.Time
		to       time.Time
		expected string
	}{
		{
			expected: ``,
		},
		{
			from:     time.Date(2022, time.January, 21, 0, 0, 0, 0, time.UTC),
			expected: `WHERE time >= '2022-01-21T00:00:00Z'`,
		},
		{
			to:       time.Date(2022, time.January, 21, 0, 0, 0, 0, time.UTC),
			expected: `WHERE time <= '2022-01-21T00:00:00Z'`,
		},
		{
			from:     time.Date(2022, time.January, 21, 0, 0, 0, 0, time.UTC),
			to:       time.Date(2022, time.January, 21, 0, 0, 0, 0, time.UTC),
			expected: `WHERE time >= '2022-01-21T00:00:00Z' AND time <= '2022-01-21T00:00:00Z'`,
		},
	}

	for _, testCase := range testCases {
		clause := makeTimestampClause(testCase.from, testCase.to)
		assert.Equal(t, testCase.expected, clause)
	}
}
