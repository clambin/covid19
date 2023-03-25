package retry

import (
	"golang.org/x/exp/constraints"
	"math"
	"time"
)

var _ BackOff = &Constant{}

type basicBackoff struct {
	MaxTries int
	current  int
}

func (b *basicBackoff) GetNext() time.Duration {
	if b.current >= b.MaxTries {
		return Stop
	}
	b.current++
	return 0
}

type Constant struct {
	basicBackoff
	Delay time.Duration
}

func (l *Constant) GetNext() time.Duration {
	if l.basicBackoff.GetNext() == Stop {
		return Stop
	}
	return l.Delay
}

func NewConstantBackoff(maxTries int, delay time.Duration) *Constant {
	return &Constant{
		basicBackoff: basicBackoff{
			MaxTries: maxTries,
		},
		Delay: delay,
	}
}

var _ BackOff = &Linear{}

type Linear struct {
	basicBackoff
	Delay    time.Duration
	MaxDelay time.Duration
}

func (l *Linear) GetNext() time.Duration {
	if l.basicBackoff.GetNext() == Stop {
		return Stop
	}
	return capped(l.Delay*time.Duration(l.current), l.MaxDelay)
}

func NewLinearBackoff(maxTries int, delay, maxDelay time.Duration) *Linear {
	return &Linear{
		basicBackoff: basicBackoff{MaxTries: maxTries},
		Delay:        delay,
		MaxDelay:     maxDelay,
	}
}

var _ BackOff = &Doubler{}

type Doubler struct {
	basicBackoff
	Delay    time.Duration
	MaxDelay time.Duration
}

func (e *Doubler) GetNext() time.Duration {
	if e.basicBackoff.GetNext() == Stop {
		return Stop
	}
	return capped(time.Duration(math.Pow(2, float64(e.current-1)))*e.Delay, e.MaxDelay)
}

func capped[T constraints.Ordered](t1, t2 T) T {
	if t1 < t2 {
		return t1
	}
	return t2
}

func NewDoublerBackoff(maxTries int, delay, maxDelay time.Duration) *Doubler {
	return &Doubler{
		basicBackoff: basicBackoff{MaxTries: maxTries},
		Delay:        delay,
		MaxDelay:     maxDelay,
	}
}
