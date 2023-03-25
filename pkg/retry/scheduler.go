package retry

import (
	"golang.org/x/exp/constraints"
	"time"
)

var _ Scheduler = &Constant{}

type Constant struct {
	MaxRetry int
	Delay    time.Duration
	current  int
}

func (l *Constant) GetNext() time.Duration {
	if l.current >= l.MaxRetry {
		return Stop
	}
	l.current++
	return l.Delay
}

var _ Scheduler = &Linear{}

type Linear struct {
	MaxRetry int
	Delay    time.Duration
	MaxDelay time.Duration
	current  int
}

func (l *Linear) GetNext() time.Duration {
	if l.current >= l.MaxRetry {
		return Stop
	}
	l.current++
	return capped(l.Delay*time.Duration(l.current), l.MaxDelay)
}

var _ Scheduler = &Doubler{}

type Doubler struct {
	MaxRetry int
	Delay    time.Duration
	MaxDelay time.Duration
	current  int
}

func (e *Doubler) GetNext() time.Duration {
	if e.current >= e.MaxRetry {
		return Stop
	}
	e.current++
	if e.current > 1 {
		e.Delay *= 2
	}
	return capped(e.Delay, e.MaxDelay)
}

func capped[T constraints.Ordered](t1, t2 T) T {
	if t1 < t2 {
		return t1
	}
	return t2
}
