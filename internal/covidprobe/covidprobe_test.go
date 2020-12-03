package covidprobe

import (
	"testing"
	"github.com/stretchr/testify/assert"
)


func TestCovidProbe(t *testing.T) {
	probe := NewCovidProbe(nil, "", "")

	probe.Run()
	assert.Equal(t, 1, probe.Measured())
	return
}

