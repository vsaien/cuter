package syncx

import (
	"time"

	"github.com/vsaien/cuter/lib/lang"
)

type Cond struct {
	signal chan lang.PlaceholderType
}

func NewCond() *Cond {
	return &Cond{
		signal: make(chan lang.PlaceholderType),
	}
}

// WaitWithTimeout wait for signal return remain wait time or timed out
func (cond *Cond) WaitWithTimeout(timeout time.Duration) (time.Duration, bool) {
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	begin := time.Now().UnixNano()
	select {
	case <-cond.signal:
		end := time.Now().UnixNano()
		remainTimeout := timeout - time.Duration(end-begin)
		return remainTimeout, true
	case <-timer.C:
		return 0, false
	}
}

// Wait for signal
func (cond *Cond) Wait() {
	<-cond.signal
}

// Signal wakes one goroutine waiting on c, if there is any.
func (cond *Cond) Signal() {
	select {
	case cond.signal <- lang.Placeholder:
	default:
	}
}
