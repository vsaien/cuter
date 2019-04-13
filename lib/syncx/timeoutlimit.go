package syncx

import (
	"errors"
	"time"
)

var ErrTimeout = errors.New("borrow timeout")

type TimeoutLimit struct {
	limit Limit
	cond  *Cond
}

func NewTimeoutLimit(n int) TimeoutLimit {
	return TimeoutLimit{
		limit: NewLimit(n),
		cond:  NewCond(),
	}
}

func (l TimeoutLimit) Borrow(timeout time.Duration) error {
	if l.TryBorrow() {
		return nil
	}

	for {
		remainTimeout, ok := l.cond.WaitWithTimeout(timeout)
		if ok && l.TryBorrow() {
			return nil
		}

		if remainTimeout <= 0 {
			return ErrTimeout
		}
	}
}

func (l TimeoutLimit) Return() error {
	if err := l.limit.Return(); err != nil {
		return err
	}

	l.cond.Signal()
	return nil
}

func (l TimeoutLimit) TryBorrow() bool {
	return l.limit.TryBorrow()
}
