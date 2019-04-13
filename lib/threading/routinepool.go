package threading

import "time"

const (
	idleTime        = 5 * time.Minute
	maxIdleRoutines = 4
)

type (
	Task func()

	PoolOption func(pool *RoutinePool)

	RoutinePool struct {
		workerPool chan chan Task
		maxIdle    int
	}
)

func NewRoutinePool(opts ...PoolOption) *RoutinePool {
	pool := &RoutinePool{
		maxIdle: maxIdleRoutines,
	}

	for _, opt := range opts {
		opt(pool)
	}

	pool.workerPool = make(chan chan Task, pool.maxIdle)

	return pool
}

func (rp *RoutinePool) Run(task Task) {
	select {
	case taskChannel := <-rp.workerPool:
		taskChannel <- task
	default:
		worker := newWorker(rp.workerPool)
		worker.startWith(task)
	}
}

func WithMaxIdle(idles int) PoolOption {
	return func(pool *RoutinePool) {
		pool.maxIdle = idles
	}
}

type worker struct {
	workerPool  chan chan Task
	taskChannel chan Task
}

func newWorker(workerPool chan chan Task) *worker {
	return &worker{
		workerPool:  workerPool,
		taskChannel: make(chan Task),
	}
}

func (w *worker) startWith(task Task) {
	go func() {
		RunSafe(task)

		timer := time.NewTimer(idleTime)
		defer timer.Stop()

		for {
			timer.Reset(idleTime)

			select {
			case w.workerPool <- w.taskChannel:
				RunSafe(<-w.taskChannel)
			case <-timer.C:
				return
			}
		}
	}()
}
