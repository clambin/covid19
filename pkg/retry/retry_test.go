package retry_test

import (
	"context"
	"errors"
	"fmt"
	"github.com/clambin/covid19/pkg/retry"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestRetry_Do(t *testing.T) {
	r := retry.Retry{BackOff: retry.NewConstantBackoff(4, 100*time.Millisecond)}

	var count int
	err := r.Do(func() error {
		count++
		return errors.New("error")
	})
	assert.Error(t, err)
	assert.Equal(t, 5, count)
}

func TestRetry_Do_WithShouldRetry(t *testing.T) {
	r := retry.Retry{
		BackOff: retry.NewConstantBackoff(4, 100*time.Millisecond),
		ShouldRetry: func(err error) bool {
			return err.Error() != "error 2"
		},
	}

	var count int
	err := r.Do(func() error {
		count++
		return fmt.Errorf("error %d", count)
	})
	assert.Error(t, err)
	assert.Equal(t, 2, count)
}

func TestRetry_Do_Success(t *testing.T) {
	r := retry.Retry{BackOff: retry.NewConstantBackoff(4, 100*time.Millisecond)}

	var count int
	err := r.Do(func() error {
		count++
		if count == 2 {
			return nil
		}
		return errors.New("error")
	})
	assert.NoError(t, err)
	assert.Equal(t, 2, count)
}

func TestRetry_DoWithContext(t *testing.T) {
	r := retry.Retry{BackOff: retry.NewConstantBackoff(4, 100*time.Millisecond)}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	var count int
	err := r.DoWithContext(ctx, func() error {
		count++
		return errors.New("error")
	})
	assert.Error(t, err)
	assert.Equal(t, "context deadline exceeded", err.Error())
	assert.NotEmpty(t, 5, count)
}
