package breaker

import "sync"

var (
	lock     sync.RWMutex
	breakers = make(map[string]Breaker)
)

func Do(name string, req func() error) error {
	return do(name, func(b Breaker) error {
		return b.Do(req)
	})
}

func DoWithAcceptable(name string, req func() error, acceptable Acceptable) error {
	return do(name, func(b Breaker) error {
		return b.DoWithAcceptable(req, acceptable)
	})
}

func DoWithFallback(name string, req func() error, fallback func(err error) error) error {
	return do(name, func(b Breaker) error {
		return b.DoWithFallback(req, fallback)
	})
}

func DoWithFallbackAcceptable(name string, req func() error, fallback func(err error) error,
	acceptable Acceptable) error {
	return do(name, func(b Breaker) error {
		return b.DoWithFallbackAcceptable(req, fallback, acceptable)
	})
}

func NoBreakFor(name string) {
	lock.Lock()
	breakers[name] = NewNoOpBreaker()
	lock.Unlock()
}

func SetBreaker(st Settings) {
	lock.Lock()
	breakers[st.Name] = NewBreakerWithSettings(st)
	lock.Unlock()
}

func do(name string, execute func(b Breaker) error) error {
	lock.RLock()
	b, ok := breakers[name]
	lock.RUnlock()
	if ok {
		return execute(b)
	} else {
		lock.Lock()
		b, ok = breakers[name]
		if ok {
			lock.Unlock()
			return execute(b)
		} else {
			b = NewBreaker()
			breakers[name] = b
			lock.Unlock()
			return execute(b)
		}
	}
}
