package breaker

import (
	"errors"
	"fmt"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const (
	customInterval = time.Second * 30
	customTimeout  = time.Second * 90
	logicFailure   = "logic failure"
)

var (
	defaultCB   *CircuitBreaker
	customCB    *CircuitBreaker
	stateChange StateChange
)

type StateChange struct {
	name string
	from State
	to   State
}

func pseudoSleep(cb *CircuitBreaker, period time.Duration) {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	if !cb.expiry.IsZero() {
		cb.expiry = cb.expiry.Add(-period)
	}
}

func succeed(cb *CircuitBreaker) error {
	return cb.Do(func() error {
		return nil
	})
}

func succeedLater(cb *CircuitBreaker, delay time.Duration) <-chan error {
	ch := make(chan error)
	go func() {
		ch <- cb.Do(func() error {
			time.Sleep(delay)
			return nil
		})
	}()
	return ch
}

func succeed2Step(cb *TwoStepCircuitBreaker) error {
	done, err := cb.Allow()
	if err != nil {
		return err
	}

	done(true)
	return nil
}

func fail(cb *CircuitBreaker) error {
	msg := "fail"
	err := cb.Do(func() error {
		return fmt.Errorf(msg)
	})
	if err.Error() == msg {
		return nil
	}
	return err
}

func fail2Step(cb *TwoStepCircuitBreaker) error {
	done, err := cb.Allow()
	if err != nil {
		return err
	}

	done(false)
	return nil
}

func causePanic(cb *CircuitBreaker) error {
	return cb.Do(func() error {
		panic("oops")
		return nil
	})
}

func logicFail(cb *CircuitBreaker) error {
	return cb.DoWithAcceptable(func() error {
		return errors.New(logicFailure)
	}, func(err error) bool {
		return err == nil || err.Error() == logicFailure
	})
}

func state(cb *CircuitBreaker) State {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	currentState, _ := cb.currentState(time.Now())
	return currentState
}

func newCustom() *CircuitBreaker {
	var customSt Settings
	customSt.Name = "cb"
	customSt.MaxRequests = 3
	customSt.Interval = customInterval
	customSt.Timeout = customTimeout
	customSt.ReadyToTrip = func(counts Counts) bool {
		numReqs := counts.Requests
		failureRatio := float64(counts.TotalFailures) / float64(numReqs)
		counts.clear() // no effect on customCB.counts

		return numReqs >= 3 && failureRatio >= 0.6
	}
	customSt.OnStateChange = func(name string, from State, to State) {
		stateChange = StateChange{name, from, to}
	}

	return NewBreakerWithSettings(customSt)
}

func init() {
	defaultCB = NewBreaker()
	customCB = newCustom()
}

func TestCircuitBreakerWithCheck(t *testing.T) {
	cb := NewBreaker()
	for i := 0; i < (defaultConsecutiveFailuresToBreak << 1); i++ {
		logicFail(cb)
	}

	assert.Equal(t, StateClosed, cb.state)
}

func TestNewCircuitBreaker(t *testing.T) {
	defaultCB := NewBreaker()

	assert.Equal(t, "", defaultCB.name)
	assert.Equal(t, uint32(defaultMaxRequests), defaultCB.maxRequests)
	assert.Equal(t, defaultInterval, defaultCB.interval)
	assert.Equal(t, defaultTimeout, defaultCB.timeout)
	assert.NotNil(t, defaultCB.readyToTrip)
	assert.Nil(t, defaultCB.onStateChange)
	assert.Equal(t, StateClosed, defaultCB.state)
	assert.Equal(t, Counts{0, 0, 0, 0, 0}, defaultCB.counts)
	assert.False(t, defaultCB.expiry.IsZero())

	customCB := newCustom()

	assert.Equal(t, "cb", customCB.name)
	assert.Equal(t, uint32(3), customCB.maxRequests)
	assert.Equal(t, time.Duration(30)*time.Second, customCB.interval)
	assert.Equal(t, time.Duration(90)*time.Second, customCB.timeout)
	assert.NotNil(t, customCB.readyToTrip)
	assert.NotNil(t, customCB.onStateChange)
	assert.Equal(t, StateClosed, customCB.state)
	assert.Equal(t, Counts{0, 0, 0, 0, 0}, customCB.counts)
	assert.False(t, customCB.expiry.IsZero())
}

func TestDefaultCircuitBreaker(t *testing.T) {
	assert.Equal(t, "", defaultCB.Name())

	for i := 0; i < 5; i++ {
		assert.Nil(t, fail(defaultCB))
	}
	assert.Equal(t, StateClosed, state(defaultCB))
	assert.Equal(t, Counts{5, 0, 5, 0, 5}, defaultCB.counts)

	assert.Nil(t, succeed(defaultCB))
	assert.Equal(t, StateClosed, state(defaultCB))
	assert.Equal(t, Counts{6, 1, 5, 1, 0}, defaultCB.counts)

	assert.Nil(t, fail(defaultCB))
	assert.Equal(t, StateClosed, state(defaultCB))
	assert.Equal(t, Counts{7, 1, 6, 0, 1}, defaultCB.counts)

	// StateClosed to StateOpen
	for i := 0; i < defaultConsecutiveFailuresToBreak; i++ {
		assert.Nil(t, fail(defaultCB)) // 6 consecutive failures
	}
	assert.Equal(t, StateOpen, state(defaultCB))
	assert.Equal(t, Counts{0, 0, 0, 0, 0}, defaultCB.counts)
	assert.False(t, defaultCB.expiry.IsZero())

	assert.Error(t, succeed(defaultCB))
	assert.Error(t, fail(defaultCB))
	assert.Equal(t, Counts{0, 0, 0, 0, 0}, defaultCB.counts)

	pseudoSleep(defaultCB, time.Duration(9)*time.Second)
	assert.Equal(t, StateOpen, state(defaultCB))

	// StateOpen to StateHalfOpen
	pseudoSleep(defaultCB, time.Duration(1)*time.Second) // over Timeout
	assert.Equal(t, StateHalfOpen, state(defaultCB))
	assert.True(t, defaultCB.expiry.IsZero())

	// StateHalfOpen to StateOpen
	assert.Nil(t, fail(defaultCB))
	assert.Equal(t, StateOpen, state(defaultCB))
	assert.Equal(t, Counts{0, 0, 0, 0, 0}, defaultCB.counts)
	assert.False(t, defaultCB.expiry.IsZero())

	// StateOpen to StateHalfOpen
	pseudoSleep(defaultCB, defaultTimeout)
	assert.Equal(t, StateHalfOpen, state(defaultCB))
	assert.True(t, defaultCB.expiry.IsZero())

	// StateHalfOpen to StateClosed
	for i := 0; i < defaultMaxRequests; i++ {
		assert.Nil(t, succeed(defaultCB))
	}
	assert.Equal(t, StateClosed, state(defaultCB))
	assert.Equal(t, Counts{0, 0, 0, 0, 0}, defaultCB.counts)
	assert.False(t, defaultCB.expiry.IsZero())
}

func TestDefaultCircuitBreakerWithFallback(t *testing.T) {
	assert.Equal(t, "", defaultCB.Name())
	verify := func(expect int) {
		var v int
		defaultCB.DoWithFallback(func() error {
			v = 1
			return nil
		}, func(err error) error {
			v = 2
			return nil
		})
		assert.Equal(t, expect, v)
	}

	verify(1)

	// StateClosed to StateOpen
	fail(defaultCB)
	for i := 0; i < defaultConsecutiveFailuresToBreak; i++ {
		assert.Nil(t, fail(defaultCB)) // 6 consecutive failures
	}
	assert.Equal(t, StateOpen, state(defaultCB))
	assert.Equal(t, Counts{0, 0, 0, 0, 0}, defaultCB.counts)
	assert.False(t, defaultCB.expiry.IsZero())

	verify(2)
}

func TestCustomCircuitBreaker(t *testing.T) {
	assert.Equal(t, "cb", customCB.Name())

	for i := 0; i < 5; i++ {
		assert.Nil(t, succeed(customCB))
		assert.Nil(t, fail(customCB))
	}
	assert.Equal(t, StateClosed, state(customCB))
	assert.Equal(t, Counts{10, 5, 5, 0, 1}, customCB.counts)

	pseudoSleep(customCB, time.Second*9)
	assert.Nil(t, succeed(customCB))
	assert.Equal(t, StateClosed, state(customCB))
	assert.Equal(t, Counts{11, 6, 5, 1, 0}, customCB.counts)

	pseudoSleep(customCB, customInterval) // over Interval
	assert.Nil(t, fail(customCB))
	assert.Equal(t, StateClosed, state(customCB))
	assert.Equal(t, Counts{1, 0, 1, 0, 1}, customCB.counts)

	// StateClosed to StateOpen
	assert.Nil(t, succeed(customCB))
	assert.Nil(t, fail(customCB)) // failure ratio: 2/3 >= 0.6
	assert.Equal(t, StateOpen, state(customCB))
	assert.Equal(t, Counts{0, 0, 0, 0, 0}, customCB.counts)
	assert.False(t, customCB.expiry.IsZero())
	assert.Equal(t, StateChange{"cb", StateClosed, StateOpen}, stateChange)

	// StateOpen to StateHalfOpen
	pseudoSleep(customCB, customTimeout)
	assert.Equal(t, StateHalfOpen, state(customCB))
	assert.False(t, defaultCB.expiry.IsZero())
	assert.Equal(t, StateChange{"cb", StateOpen, StateHalfOpen}, stateChange)

	assert.Nil(t, succeed(customCB))
	assert.Nil(t, succeed(customCB))
	assert.Equal(t, StateHalfOpen, state(customCB))
	assert.Equal(t, Counts{2, 2, 0, 2, 0}, customCB.counts)

	// StateHalfOpen to StateClosed
	ch := succeedLater(customCB, time.Duration(100)*time.Millisecond) // 3 consecutive successes
	time.Sleep(time.Duration(50) * time.Millisecond)
	customCB.mutex.Lock()
	assert.Equal(t, Counts{3, 2, 0, 2, 0}, customCB.counts)
	customCB.mutex.Unlock()
	assert.Error(t, succeed(customCB)) // over MaxRequests
	assert.Nil(t, <-ch)
	assert.Equal(t, StateClosed, state(customCB))
	assert.Equal(t, Counts{0, 0, 0, 0, 0}, customCB.counts)
	assert.False(t, customCB.expiry.IsZero())
	assert.Equal(t, StateChange{"cb", StateHalfOpen, StateClosed}, stateChange)
}

func TestTwoStepCircuitBreaker(t *testing.T) {
	tscb := NewTwoStepBreakerWithSettings(Settings{Name: "tscb"})
	assert.Equal(t, "tscb", tscb.Name())

	for i := 0; i < 5; i++ {
		assert.Nil(t, fail2Step(tscb))
	}

	assert.Equal(t, StateClosed, state(tscb.cb))
	assert.Equal(t, Counts{5, 0, 5, 0, 5}, tscb.cb.counts)

	assert.Nil(t, succeed2Step(tscb))
	assert.Equal(t, StateClosed, state(tscb.cb))
	assert.Equal(t, Counts{6, 1, 5, 1, 0}, tscb.cb.counts)

	assert.Nil(t, fail2Step(tscb))
	assert.Equal(t, StateClosed, state(tscb.cb))
	assert.Equal(t, Counts{7, 1, 6, 0, 1}, tscb.cb.counts)

	// StateClosed to StateOpen
	for i := 0; i < defaultConsecutiveFailuresToBreak; i++ {
		assert.Nil(t, fail2Step(tscb)) // 6 consecutive failures
	}
	assert.Equal(t, StateOpen, state(tscb.cb))
	assert.Equal(t, Counts{0, 0, 0, 0, 0}, tscb.cb.counts)
	assert.False(t, tscb.cb.expiry.IsZero())

	assert.Error(t, succeed2Step(tscb))
	assert.Error(t, fail2Step(tscb))
	assert.Equal(t, Counts{0, 0, 0, 0, 0}, tscb.cb.counts)

	pseudoSleep(tscb.cb, time.Second*9)
	assert.Equal(t, StateOpen, state(tscb.cb))

	// StateOpen to StateHalfOpen
	pseudoSleep(tscb.cb, time.Second) // over Timeout
	assert.Equal(t, StateHalfOpen, state(tscb.cb))
	assert.True(t, tscb.cb.expiry.IsZero())

	// StateHalfOpen to StateOpen
	assert.Nil(t, fail2Step(tscb))
	assert.Equal(t, StateOpen, state(tscb.cb))
	assert.Equal(t, Counts{0, 0, 0, 0, 0}, tscb.cb.counts)
	assert.False(t, tscb.cb.expiry.IsZero())

	// StateOpen to StateHalfOpen
	pseudoSleep(tscb.cb, defaultTimeout)
	assert.Equal(t, StateHalfOpen, state(tscb.cb))
	assert.True(t, tscb.cb.expiry.IsZero())

	// StateHalfOpen to StateClosed
	for i := 0; i < defaultMaxRequests; i++ {
		assert.Nil(t, succeed2Step(tscb))
	}
	assert.Equal(t, StateClosed, state(tscb.cb))
	assert.Equal(t, Counts{0, 0, 0, 0, 0}, tscb.cb.counts)
	assert.False(t, tscb.cb.expiry.IsZero())
}

func TestPanicInRequest(t *testing.T) {
	cb := NewBreaker()
	assert.Panics(t, func() { causePanic(cb) })
	assert.Equal(t, Counts{1, 0, 1, 0, 1}, cb.counts)
}

func TestGeneration(t *testing.T) {
	pseudoSleep(customCB, time.Second*4)
	assert.Nil(t, succeed(customCB))
	ch := succeedLater(customCB, time.Millisecond*500)
	time.Sleep(time.Millisecond * 400)
	customCB.mutex.Lock()
	assert.Equal(t, Counts{2, 1, 0, 1, 0}, customCB.counts)
	customCB.mutex.Unlock()

	pseudoSleep(customCB, time.Second*30)
	assert.Equal(t, StateClosed, state(customCB))
	assert.Equal(t, Counts{0, 0, 0, 0, 0}, customCB.counts)

	// the request from the previous generation has no effect on customCB.counts
	assert.Nil(t, <-ch)
	assert.Equal(t, Counts{0, 0, 0, 0, 0}, customCB.counts)
}

func TestCircuitBreakerInParallel(t *testing.T) {
	runtime.GOMAXPROCS(runtime.NumCPU())

	ch := make(chan error)

	const numReqs = 10000
	routine := func() {
		for i := 0; i < numReqs; i++ {
			ch <- succeed(customCB)
		}
	}

	const numRoutines = 10
	for i := 0; i < numRoutines; i++ {
		go routine()
	}

	total := uint32(numReqs * numRoutines)
	for i := uint32(0); i < total; i++ {
		err := <-ch
		assert.Nil(t, err)
	}
	assert.Equal(t, Counts{total, total, 0, total, 0}, customCB.counts)
}
