package syncx

import "sync"

type (
	ExclusiveCalls interface {
		Do(key string, fn func() (interface{}, error)) (interface{}, error)
		DoEx(key string, fn func() (interface{}, error)) (interface{}, bool, error)
	}

	call struct {
		wg  sync.WaitGroup
		val interface{}
		err error
	}

	exclusiveGroup struct {
		mu sync.Mutex
		m  map[string]*call
	}
)

func NewExclusiveCalls() ExclusiveCalls {
	return &exclusiveGroup{
		m: make(map[string]*call),
	}
}

func (eg *exclusiveGroup) Do(key string, fn func() (interface{}, error)) (interface{}, error) {
	eg.mu.Lock()
	if c, ok := eg.m[key]; ok {
		eg.mu.Unlock()
		c.wg.Wait()
		return c.val, c.err
	}

	c := eg.makeCall(key, fn)
	return c.val, c.err
}

func (eg *exclusiveGroup) DoEx(key string, fn func() (interface{}, error)) (val interface{}, fresh bool, err error) {
	eg.mu.Lock()
	if c, ok := eg.m[key]; ok {
		eg.mu.Unlock()
		c.wg.Wait()
		return c.val, false, c.err
	}

	c := eg.makeCall(key, fn)
	return c.val, true, c.err
}

func (eg *exclusiveGroup) makeCall(key string, fn func() (interface{}, error)) *call {
	c := new(call)
	c.wg.Add(1)
	eg.m[key] = c
	eg.mu.Unlock()

	defer func() {
		// delete key first, done later. can't reverse the order, because if reverse,
		// another Do call might wg.Wait() without get notified with wg.Done()
		eg.mu.Lock()
		delete(eg.m, key)
		eg.mu.Unlock()
		c.wg.Done()
	}()

	c.val, c.err = fn()
	return c
}
