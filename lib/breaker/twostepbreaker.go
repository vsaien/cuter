package breaker

// TwoStepCircuitBreaker is like CircuitBreaker but instead of surrounding a function
// with the breaker functionality, it only checks whether a request can proceed and
// expects the caller to report the outcome in a separate step using a callback.
type TwoStepCircuitBreaker struct {
	cb *CircuitBreaker
}

// NewTwoStepBreaker returns a new TwoStepCircuitBreaker with default settings.
func NewTwoStepBreaker() *TwoStepCircuitBreaker {
	return &TwoStepCircuitBreaker{
		cb: NewBreaker(),
	}
}

// NewTwoStepBreakerWithSettings returns a new TwoStepCircuitBreaker configured with the given Settings.
func NewTwoStepBreakerWithSettings(st Settings) *TwoStepCircuitBreaker {
	return &TwoStepCircuitBreaker{
		cb: NewBreakerWithSettings(st),
	}
}

// Name returns the name of the TwoStepCircuitBreaker.
func (tcb *TwoStepCircuitBreaker) Name() string {
	return tcb.cb.Name()
}

// Allow checks if a new request can proceed. It returns a callback that should be used to
// register the success or failure in a separate step. If the circuit breaker doesn't allow
// requests, it returns an error.
func (tcb *TwoStepCircuitBreaker) Allow() (done func(success bool), err error) {
	generation, err := tcb.cb.beforeRequest()
	if err != nil {
		return nil, err
	}

	return func(success bool) {
		tcb.cb.afterRequest(generation, success)
	}, nil
}
