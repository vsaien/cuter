package syncx

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTimeoutCondWait(t *testing.T) {
	var wait sync.WaitGroup
	cond := NewCond()
	wait.Add(2)
	go func() {
		cond.Wait()
		wait.Done()
	}()
	time.Sleep(time.Duration(50) * time.Millisecond)
	go func() {
		cond.Signal()
		wait.Done()
	}()
	wait.Wait()
}

func TestTimeoutCondWaitTimeout(t *testing.T) {
	var wait sync.WaitGroup
	cond := NewCond()
	wait.Add(1)
	go func() {
		cond.WaitWithTimeout(time.Duration(500) * time.Millisecond)
		wait.Done()
	}()
	wait.Wait()
}

func TestTimeoutCondWaitTimeoutNotify(t *testing.T) {
	var wait sync.WaitGroup
	cond := NewCond()
	wait.Add(2)
	ch := make(chan int, 1)
	timeout := 2000
	go func() {
		begin := currentTimeMillis()
		cond.WaitWithTimeout(time.Duration(timeout) * time.Millisecond)
		end := currentTimeMillis()
		ch <- int(end - begin)
		wait.Done()
	}()
	sleep(200)
	go func() {
		cond.Signal()
		wait.Done()
	}()
	wait.Wait()
	time := <-ch
	close(ch)
	assert.True(t, time < timeout)
	assert.True(t, time >= 200)
}

func sleep(millisecond int) {
	time.Sleep(time.Duration(millisecond) * time.Millisecond)
}

func currentTimeMillis() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

func TestTimeoutCondWaitTimeoutRemain(t *testing.T) {
	var wait sync.WaitGroup
	cond := NewCond()
	wait.Add(2)
	ch := make(chan time.Duration, 1)
	timeout := time.Duration(2000) * time.Millisecond
	go func() {
		remainTimeout, _ := cond.WaitWithTimeout(timeout)
		ch <- remainTimeout
		wait.Done()
	}()
	sleep(200)
	go func() {
		cond.Signal()
		wait.Done()
	}()
	wait.Wait()
	remainTimeout := <-ch
	close(ch)
	assert.True(t, remainTimeout < timeout, "expect remainTimeout %v < %v", remainTimeout, timeout)
	assert.True(t, remainTimeout >= time.Duration(200)*time.Millisecond,
		"expect remainTimeout %v >= 200 millisecond", remainTimeout)
}

func TestSignalNoWait(t *testing.T) {
	cond := NewCond()
	cond.Signal()
}
