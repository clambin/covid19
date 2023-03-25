package retry_test

import (
	"github.com/clambin/covid19/pkg/retry"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestConstant(t *testing.T) {
	s := retry.Constant{Delay: time.Second, MaxRetry: 5}

	delays := []time.Duration{
		time.Second,
		time.Second,
		time.Second,
		time.Second,
		time.Second,
		retry.Stop,
	}

	for _, delay := range delays {
		newDelay := s.GetNext()
		assert.Equal(t, delay, newDelay)
	}
}

func TestLinear(t *testing.T) {
	s := retry.Linear{Delay: time.Second, MaxRetry: 5, MaxDelay: 4 * time.Second}

	delays := []time.Duration{
		time.Second,
		2 * time.Second,
		3 * time.Second,
		4 * time.Second,
		4 * time.Second,
		retry.Stop,
	}

	for _, delay := range delays {
		newDelay := s.GetNext()
		assert.Equal(t, delay, newDelay)
	}
}

func TestDoubler(t *testing.T) {
	s := retry.Doubler{Delay: time.Second, MaxRetry: 5, MaxDelay: 8 * time.Second}

	delays := []time.Duration{
		time.Second,
		2 * time.Second,
		4 * time.Second,
		8 * time.Second,
		8 * time.Second,
		retry.Stop,
	}

	for _, delay := range delays {
		newDelay := s.GetNext()
		assert.Equal(t, delay, newDelay)
	}
}
