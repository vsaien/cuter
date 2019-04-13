package syncx

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/vsaien/cuter/lib/lang"

	"github.com/stretchr/testify/assert"
)

const limit = 10

func TestPoolGet(t *testing.T) {
	stack := NewPool(limit, create, destroy)
	ch := make(chan lang.PlaceholderType)

	for i := 0; i < limit; i++ {
		go func() {
			v := stack.Get()
			if v.(int) != 1 {
				t.Fatal("unmatch value")
			}
			ch <- lang.Placeholder
		}()

		select {
		case <-ch:
		case <-time.After(time.Millisecond):
			t.Fail()
		}
	}
}

func TestPoolPopTooMany(t *testing.T) {
	stack := NewPool(limit, create, destroy)
	ch := make(chan lang.PlaceholderType, 1)

	for i := 0; i < limit; i++ {
		var wait sync.WaitGroup
		wait.Add(1)
		go func() {
			stack.Get()
			ch <- lang.Placeholder
			wait.Done()
		}()

		wait.Wait()
		select {
		case <-ch:
		default:
			t.Fail()
		}
	}

	var waitGroup, pushWait sync.WaitGroup
	waitGroup.Add(1)
	pushWait.Add(1)
	go func() {
		pushWait.Done()
		stack.Get()
		waitGroup.Done()
	}()

	pushWait.Wait()
	stack.Put(1)
	waitGroup.Wait()
}

func TestPoolPopFirst(t *testing.T) {
	var value int32
	stack := NewPool(limit, func() interface{} {
		return atomic.AddInt32(&value, 1)
	}, destroy)

	for i := 0; i < 100; i++ {
		v := stack.Get().(int32)
		assert.Equal(t, 1, int(v))
		stack.Put(v)
	}
}

func create() interface{} {
	return 1
}

func destroy(_ interface{}) {
}