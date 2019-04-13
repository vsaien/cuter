package collection

import (
	"sync"
	"testing"
	"time"

	"github.com/vsaien/cuter/lib/lang"

	"github.com/stretchr/testify/assert"
)

func TestTimingWheel(t *testing.T) {
	const (
		set = iota
		move
		remove
	)

	type (
		action struct {
			op       int
			at       int
			value    int
			duration time.Duration
			done     bool
		}

		testCase struct {
			key         interface{}
			actions     []*action
			expectTime  int
			expectValue int
		}

		kv struct {
			key   int
			value int
		}
	)

	var tests = []*testCase{
		{
			key: 1,
			actions: []*action{
				{
					op:       set,
					at:       0,
					value:    1,
					duration: time.Minute * 5,
				},
				{
					op:       move,
					at:       1,
					duration: time.Minute * 7,
				},
				{
					op:       move,
					at:       6,
					duration: time.Minute * 8,
				},
				{
					op:       move,
					at:       12,
					duration: time.Minute * 9,
				},
			},
		},
		{
			key: 2,
			actions: []*action{
				{
					op:       set,
					at:       0,
					value:    2,
					duration: time.Second * 390,
				},
				{
					op: remove,
					at: 1,
				},
			},
		},
		{
			key: 3,
			actions: []*action{
				{
					op:       set,
					at:       0,
					value:    3,
					duration: time.Minute * 11,
				},
			},
		},
		{
			key: 4,
			actions: []*action{
				{
					op:       set,
					at:       0,
					value:    4,
					duration: time.Minute * 11,
				},
				{
					op:       move,
					at:       6,
					duration: time.Minute * 2,
				},
			},
		},
		{
			key: 5,
			actions: []*action{
				{
					op:       set,
					at:       0,
					value:    1,
					duration: time.Minute * 5,
				},
				{
					op:       move,
					at:       1,
					duration: time.Minute * 7,
				},
				{
					op:       move,
					at:       6,
					duration: time.Minute * 8,
				},
				{
					op:       set,
					at:       8,
					value:    3,
					duration: time.Minute * 6,
				},
				{
					op:       move,
					at:       12,
					duration: time.Minute * 9,
				},
			},
		},
	}

	var i int
	var shadow int
	var lock sync.Mutex
	result := make(map[int]kv)
	tw, _ := newMockedTimingWheel(time.Minute, 10, func(k, v interface{}) {
		lock.Lock()
		result[k.(int)] = kv{
			key:   shadow,
			value: v.(int),
		}
		lock.Unlock()
	})

	tick := func() {
		time.Sleep(time.Millisecond * 5)
		tw.testChan <- lang.Placeholder
	}

	ticks := 1
	for ; ticks >= 0; i++ {
		lock.Lock()
		shadow = i
		lock.Unlock()
		for _, test := range tests {
			for _, act := range test.actions {
				if act.done {
					continue
				}
				if act.at > 0 {
					act.at--
				} else {
					act.done = true
					switch act.op {
					case set:
						seconds := int(act.duration.Minutes())
						if ticks < seconds {
							ticks = seconds
						}
						test.expectTime = i + seconds
						test.expectValue = act.value
						tw.SetTimer(test.key, act.value, act.duration)
					case move:
						seconds := int(act.duration.Minutes())
						if ticks < seconds {
							ticks = seconds
						}
						test.expectTime = i + seconds
						tw.MoveTimer(test.key, act.duration)
					case remove:
						test.expectTime = 0
						test.expectValue = 0
						tw.RemoveTimer(test.key)
					}
				}
			}
		}

		tick()
		ticks--
	}

	for _, test := range tests {
		lock.Lock()
		assert.Equal(t, test.expectTime, result[test.key.(int)].key)
		assert.Equal(t, test.expectValue, result[test.key.(int)].value)
		lock.Unlock()
	}
}

func BenchmarkTimingWheel(b *testing.B) {
	b.ReportAllocs()

	tw, _ := NewTimingWheel(time.Second, 100, func(k, v interface{}) {})
	for i := 0; i < b.N; i++ {
		tw.SetTimer(i, i, time.Second)
		tw.SetTimer(b.N+i, b.N+i, time.Second)
		tw.MoveTimer(i, time.Second*time.Duration(i))
		tw.RemoveTimer(i)
	}
}

type mockedTimingWheel struct {
	*TimingWheel
	testChan chan lang.PlaceholderType
}

func newMockedTimingWheel(interval time.Duration, numSlots int, execute Execute) (*mockedTimingWheel, error) {
	tw, err := NewTimingWheel(interval, numSlots, execute)
	if err != nil {
		return nil, err
	}

	tw.Stop()
	mtw := &mockedTimingWheel{
		TimingWheel: tw,
		testChan:    make(chan lang.PlaceholderType),
	}
	go mtw.run()

	return mtw, nil
}

func (mtw *mockedTimingWheel) run() {
	for {
		select {
		case <-mtw.testChan:
			mtw.onTick()
		case task := <-mtw.setChannel:
			mtw.setTask(&task)
		case key := <-mtw.removeChannel:
			mtw.removeTask(key)
		case task := <-mtw.moveChannel:
			mtw.moveTask(task)
		}
	}
}
