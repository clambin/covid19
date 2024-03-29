package retry

import (
	"context"
	"time"
)

type Retry struct {
	BackOff
	ShouldRetry func(error) bool
}

type BackOff interface {
	GetNext() time.Duration
}

const Stop time.Duration = -1

func (r *Retry) Do(f func() error) error {
	return r.DoWithContext(context.Background(), f)
}

func (r *Retry) DoWithContext(ctx context.Context, f func() error) error {
	for {
		err := f()
		if err == nil {
			return nil
		}

		if r.ShouldRetry != nil && !r.ShouldRetry(err) {
			return err
		}

		delay := r.BackOff.GetNext()
		if delay == Stop {
			return err
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
		}
	}
}
