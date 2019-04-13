package breaker

import (
	"errors"
	"sync"
	"time"
)

const (
	StateClosed State = iota
	StateHalfOpen
	StateOpen
)

const (
	defaultMaxRequests                = 3
	defaultConsecutiveFailuresToBreak = 5
	defaultInterval                   = time.Second * 5
	defaultTimeout                    = time.Second * 10
	noOpBreakerName                   = "NoOpBreaker"
)

var (
	// ErrTooManyRequests is returned when the CB state is half open and the requests count is over the cb maxRequests
	ErrTooManyRequests = errors.New("too many requests on half-open state")
	// ErrOpenState is returned when the CB state is open
	ErrOpenState = errors.New("circuit breaker is open")
)

// Counts holds the numbers of requests and their successes/failures.
// CircuitBreaker clears the internal Counts either
// on the change of the state or at the closed-state intervals.
// Counts ignores the results of the requests sent before clearing.
type (
	Breaker interface {
		Name() string
		Do(req func() error) error
		DoWithAcceptable(req func() error, acceptable Acceptable) error
		DoWithFallback(req func() error, fallback func(err error) error) error
		DoWithFallbackAcceptable(req func() error, fallback func(err error) error, acceptable Acceptable) error
	}

	State      int
	Acceptable func(err error) bool

	Counts struct {
		Requests             uint32
		TotalSuccesses       uint32
		TotalFailures        uint32
		ConsecutiveSuccesses uint32
		ConsecutiveFailures  uint32
	}

	// Settings configures CircuitBreaker:
	//
	// Name is the name of the CircuitBreaker.
	//
	// MaxRequests is the maximum number of requests allowed to pass through
	// when the CircuitBreaker is half-open.
	// If MaxRequests is 0, the CircuitBreaker allows only 3 request.
	//
	// Interval is the cyclic period of the closed state
	// for the CircuitBreaker to clear the internal Counts.
	// If Interval is 0, the interval value of the CircuitBreaker is set to 5 seconds.
	//
	// Timeout is the period of the open state,
	// after which the state of the CircuitBreaker becomes half-open.
	// If Timeout is 0, the timeout value of the CircuitBreaker is set to 10 seconds.
	//
	// ReadyToTrip is called with a copy of Counts whenever a request fails in the closed state.
	// If ReadyToTrip returns true, the CircuitBreaker will be placed into the open state.
	// If ReadyToTrip is nil, default ReadyToTrip is used.
	// Default ReadyToTrip returns true when the number of consecutive failures is more than 5.
	//
	// OnStateChange is called whenever the state of the CircuitBreaker changes.
	Settings struct {
		Name          string
		MaxRequests   uint32
		Interval      time.Duration
		Timeout       time.Duration
		ReadyToTrip   func(counts Counts) bool
		OnStateChange func(name string, from State, to State)
	}

	// CircuitBreaker is a state machine to prevent sending requests that are likely to fail.
	CircuitBreaker struct {
		name          string
		maxRequests   uint32
		interval      time.Duration
		timeout       time.Duration
		readyToTrip   func(counts Counts) bool
		onStateChange func(name string, from State, to State)

		mutex      sync.Mutex
		state      State
		generation uint64
		counts     Counts
		expiry     time.Time
	}

	NoOpBreaker struct {
	}
)

func (c *Counts) onRequest() {
	c.Requests++
}

func (c *Counts) onSuccess() {
	c.TotalSuccesses++
	c.ConsecutiveSuccesses++
	c.ConsecutiveFailures = 0
}

func (c *Counts) onFailure() {
	c.TotalFailures++
	c.ConsecutiveFailures++
	c.ConsecutiveSuccesses = 0
}

func (c *Counts) clear() {
	c.Requests = 0
	c.TotalSuccesses = 0
	c.TotalFailures = 0
	c.ConsecutiveSuccesses = 0
	c.ConsecutiveFailures = 0
}

// NewBreaker returns a new CircuitBreaker with default settings.
func NewBreaker() *CircuitBreaker {
	cb := new(CircuitBreaker)
	withSettings(Settings{})(cb)
	cb.toNewGeneration(time.Now())

	return cb
}

// NewBreakerWithSettings returns a new CircuitBreaker configured with the given Settings.
func NewBreakerWithSettings(st Settings) *CircuitBreaker {
	cb := new(CircuitBreaker)
	withSettings(st)(cb)
	cb.toNewGeneration(time.Now())

	return cb
}

// Name returns the name of the CircuitBreaker.
func (cb *CircuitBreaker) Name() string {
	return cb.name
}

// Do runs the given request if the CircuitBreaker accepts it.
// Do returns an error instantly if the CircuitBreaker rejects the request.
// If a panic occurs in the request, the CircuitBreaker handles it as an error
// and causes the same panic again.
func (cb *CircuitBreaker) Do(req func() error) error {
	return cb.doReq(req, nil, defaultAcceptable)
}

// DoWithAcceptable runs the given request if the CircuitBreaker accepts it.
// Do returns an error instantly if the CircuitBreaker rejects the request.
// If a panic occurs in the request, the CircuitBreaker handles it as an error
// and causes the same panic again.
// acceptable checks if it's a successful call, even if the err is not nil.
func (cb *CircuitBreaker) DoWithAcceptable(req func() error, acceptable Acceptable) error {
	return cb.doReq(req, nil, acceptable)
}

// DoWithFallback runs the given request if the CircuitBreaker accepts it.
// DoWithFallback runs the fallback if the CircuitBreaker rejects the request.
// If a panic occurs in the request, the CircuitBreaker handles it as an error
// and causes the same panic again.
func (cb *CircuitBreaker) DoWithFallback(req func() error, fallback func(err error) error) error {
	return cb.doReq(req, fallback, defaultAcceptable)
}

// DoWithFallbackAcceptable runs the given request if the CircuitBreaker accepts it.
// DoWithFallback runs the fallback if the CircuitBreaker rejects the request.
// If a panic occurs in the request, the CircuitBreaker handles it as an error
// and causes the same panic again.
// acceptable checks if it's a successful call, even if the err is not nil.
func (cb *CircuitBreaker) DoWithFallbackAcceptable(req func() error, fallback func(err error) error,
	acceptable Acceptable) error {
	return cb.doReq(req, fallback, acceptable)
}

func (cb *CircuitBreaker) doReq(req func() error, fallback func(err error) error, acceptable Acceptable) error {
	generation, err := cb.beforeRequest()
	if err != nil {
		if fallback != nil {
			return fallback(err)
		} else {
			return err
		}
	}

	defer func() {
		e := recover()
		if e != nil {
			cb.afterRequest(generation, false)
			panic(e)
		}
	}()

	err = req()
	cb.afterRequest(generation, acceptable(err))
	return err
}

func (cb *CircuitBreaker) beforeRequest() (uint64, error) {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	now := time.Now()
	state, generation := cb.currentState(now)
	if state == StateOpen {
		return generation, ErrOpenState
	} else if state == StateHalfOpen && cb.counts.Requests >= cb.maxRequests {
		return generation, ErrTooManyRequests
	}

	cb.counts.onRequest()
	return generation, nil
}

func (cb *CircuitBreaker) afterRequest(before uint64, success bool) {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	now := time.Now()
	state, generation := cb.currentState(now)
	if generation != before {
		return
	}

	if success {
		cb.onSuccess(state, now)
	} else {
		cb.onFailure(state, now)
	}
}

func (cb *CircuitBreaker) onSuccess(state State, now time.Time) {
	switch state {
	case StateClosed:
		cb.counts.onSuccess()
	case StateHalfOpen:
		cb.counts.onSuccess()
		if cb.counts.ConsecutiveSuccesses >= cb.maxRequests {
			cb.setState(StateClosed, now)
		}
	}
}

func (cb *CircuitBreaker) onFailure(state State, now time.Time) {
	switch state {
	case StateClosed:
		cb.counts.onFailure()
		if cb.readyToTrip(cb.counts) {
			cb.setState(StateOpen, now)
		}
	case StateHalfOpen:
		cb.setState(StateOpen, now)
	}
}

func (cb *CircuitBreaker) currentState(now time.Time) (State, uint64) {
	switch cb.state {
	case StateClosed:
		if !cb.expiry.IsZero() && cb.expiry.Before(now) {
			cb.toNewGeneration(now)
		}
	case StateOpen:
		if cb.expiry.Before(now) {
			cb.setState(StateHalfOpen, now)
		}
	}

	return cb.state, cb.generation
}

func (cb *CircuitBreaker) setState(state State, now time.Time) {
	if cb.state == state {
		return
	}

	prev := cb.state
	cb.state = state
	cb.toNewGeneration(now)

	if cb.onStateChange != nil {
		cb.onStateChange(cb.name, prev, state)
	}
}

func (cb *CircuitBreaker) toNewGeneration(now time.Time) {
	cb.generation++
	cb.counts.clear()

	var zero time.Time
	switch cb.state {
	case StateClosed:
		if cb.interval == 0 {
			cb.expiry = zero
		} else {
			cb.expiry = now.Add(cb.interval)
		}
	case StateOpen:
		cb.expiry = now.Add(cb.timeout)
	default: // StateHalfOpen
		cb.expiry = zero
	}
}

func NewNoOpBreaker() Breaker {
	return NoOpBreaker{}
}

func (b NoOpBreaker) Name() string {
	return noOpBreakerName
}

func (b NoOpBreaker) Do(req func() error) error {
	return req()
}

func (b NoOpBreaker) DoWithAcceptable(req func() error, acceptable Acceptable) error {
	return req()
}

func (b NoOpBreaker) DoWithFallback(req func() error, fallback func(err error) error) error {
	return req()
}

func (b NoOpBreaker) DoWithFallbackAcceptable(req func() error, fallback func(err error) error,
	acceptable Acceptable) error {
	return req()
}

func defaultAcceptable(err error) bool {
	return err == nil
}

func readyToTrip(counts Counts) bool {
	return counts.ConsecutiveFailures > defaultConsecutiveFailuresToBreak
}

func withSettings(st Settings) func(*CircuitBreaker) {
	return func(cb *CircuitBreaker) {
		cb.name = st.Name
		cb.onStateChange = st.OnStateChange

		if st.MaxRequests == 0 {
			cb.maxRequests = defaultMaxRequests
		} else {
			cb.maxRequests = st.MaxRequests
		}

		if st.Interval == 0 {
			cb.interval = defaultInterval
		} else {
			cb.interval = st.Interval
		}

		if st.Timeout == 0 {
			cb.timeout = defaultTimeout
		} else {
			cb.timeout = st.Timeout
		}

		if st.ReadyToTrip == nil {
			cb.readyToTrip = readyToTrip
		} else {
			cb.readyToTrip = st.ReadyToTrip
		}
	}
}
