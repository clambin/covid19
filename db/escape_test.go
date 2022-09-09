package db

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

type testcase struct {
	in  string
	out string
}

func TestEscapeString(t *testing.T) {
	var testcases = []testcase{
		{
			in:  "foo",
			out: "foo",
		},
		{
			in:  "my 'foo",
			out: "my ''foo",
		},
		{
			in:  "'hello'",
			out: "''hello''",
		},
	}

	for _, test := range testcases {
		assert.Equal(t, test.out, escapeString(test.in))
	}
}
