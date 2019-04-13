package threading

import (
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRoutinePool(t *testing.T) {
	times := 100
	pool := NewRoutinePool()

	var counter int32
	var waitGroup sync.WaitGroup
	for i := 0; i < times; i++ {
		waitGroup.Add(1)
		pool.Run(func() {
			atomic.AddInt32(&counter, 1)
			waitGroup.Done()
		})
	}

	waitGroup.Wait()

	assert.Equal(t, times, int(counter))
}
