package executors

import (
	"sync"
	"time"

	"github.com/vsaien/cuter/lib/lang"
	"github.com/vsaien/cuter/lib/threading"
)

type (
	// A type that satisfies executors.TaskContainer can be used as the underlying
	// container that used to do periodical executions.
	TaskContainer interface {
		// AddTask adds the task into the container.
		// Returns true if the container needs to be flushed after the addition.
		AddTask(task interface{}) bool
		// Execute handles the collected tasks by the container when flushing.
		Execute(tasks interface{})
		// RemoveAll removes the contained tasks, and return them.
		RemoveAll() interface{}
	}

	PeriodicalExecutor struct {
		commander chan lang.PlaceholderType
		interval  time.Duration
		container TaskContainer
		lock      sync.Mutex
	}
)

func NewPeriodicalExecutor(interval time.Duration, container TaskContainer) *PeriodicalExecutor {
	executor := &PeriodicalExecutor{
		commander: make(chan lang.PlaceholderType),
		interval:  interval,
		container: container,
	}
	executor.backgroundFlush()

	return executor
}

func (pe *PeriodicalExecutor) Add(task interface{}) {
	if pe.addAndCheck(task) {
		pe.commander <- lang.Placeholder
	}
}

func (pe *PeriodicalExecutor) ForceFlush() {
	pe.flushOnce()
}

func (pe *PeriodicalExecutor) Sync(fn func()) {
	pe.lock.Lock()
	defer pe.lock.Unlock()
	fn()
}

func (pe *PeriodicalExecutor) addAndCheck(task interface{}) bool {
	pe.lock.Lock()
	defer pe.lock.Unlock()
	return pe.container.AddTask(task)
}

func (pe *PeriodicalExecutor) backgroundFlush() {
	threading.GoSafe(func() {
		ticker := time.NewTicker(pe.interval)
		defer ticker.Stop()

		var commanded bool
		for {
			select {
			case <-pe.commander:
				commanded = true
				pe.flushOnce()
			case <-ticker.C:
				if commanded {
					commanded = false
				} else {
					pe.flushOnce()
				}
			}
		}
	})
}

func (pe *PeriodicalExecutor) drainCommander() {
	for {
		select {
		case <-pe.commander:
			// let goroutines that blocked on commander pass
		default:
			return
		}
	}
}

func (pe *PeriodicalExecutor) flushOnce() {
	values := pe.removeAll()
	pe.drainCommander()
	pe.container.Execute(values)
}

func (pe *PeriodicalExecutor) removeAll() interface{} {
	pe.lock.Lock()
	defer pe.lock.Unlock()
	return pe.container.RemoveAll()
}
