package traffic

import (
	"os"
	"sync"
	"time"

	"github.com/vsaien/cuter/lib/executors"
	"github.com/vsaien/cuter/lib/logx"
)

var (
	LogInterval = time.Minute

	writerLock   sync.Mutex
	reportWriter Writer = nil
)

type (
	Writer interface {
		Write(report *Report) error
	}

	Report struct {
		Name          string  `json:"name"`
		Timestamp     int64   `json:"tm"`
		Pid           int     `json:"pid"`
		ReqsPerSecond float32 `json:"tps"`
		Average       float32 `json:"avg"`
		Median        float32 `json:"med"`
		Top90th       float32 `json:"t90"`
		Top99th       float32 `json:"t99"`
		Top99p9th     float32 `json:"t99p9"`
	}

	Metrics struct {
		executor  *executors.PeriodicalExecutor
		container *metricsContainer
	}
)

func SetReportWriter(writer Writer) {
	writerLock.Lock()
	reportWriter = writer
	writerLock.Unlock()
}

func NewMetrics(name string) *Metrics {
	container := &metricsContainer{
		name: name,
		pid:  os.Getpid(),
	}

	return &Metrics{
		executor:  executors.NewPeriodicalExecutor(LogInterval, container),
		container: container,
	}
}

func (m *Metrics) Add(task Task) {
	m.executor.Add(task)
}

func (m *Metrics) SetName(name string) {
	m.executor.Sync(func() {
		m.container.name = name
	})
}

type (
	tasksDurationPair struct {
		tasks    []Task
		duration time.Duration
	}

	metricsContainer struct {
		name     string
		pid      int
		tasks    []Task
		duration time.Duration
	}
)

func (c *metricsContainer) AddTask(v interface{}) bool {
	if task, ok := v.(Task); ok {
		c.tasks = append(c.tasks, task)
		c.duration += task.Duration
	}

	return false
}

func (c *metricsContainer) Execute(v interface{}) {
	pair := v.(tasksDurationPair)
	tasks := pair.tasks
	duration := pair.duration
	size := len(tasks)
	report := &Report{
		Name:          c.name,
		Timestamp:     time.Now().Unix(),
		Pid:           c.pid,
		ReqsPerSecond: float32(size) / float32(LogInterval/time.Second),
	}

	if size > 0 {
		report.Average = float32(duration/time.Millisecond) / float32(size)

		fiftyPercent := size >> 1
		if fiftyPercent > 0 {
			top50pTasks := topK(tasks, fiftyPercent)
			medianTask := top50pTasks[0]
			report.Median = float32(medianTask.Duration) / float32(time.Millisecond)
			tenPercent := fiftyPercent / 5
			if tenPercent > 0 {
				top10pTasks := topK(tasks, tenPercent)
				task90th := top10pTasks[0]
				report.Top90th = float32(task90th.Duration) / float32(time.Millisecond)
				onePercent := tenPercent / 10
				if onePercent > 0 {
					top1pTasks := topK(top10pTasks, onePercent)
					task99th := top1pTasks[0]
					report.Top99th = float32(task99th.Duration) / float32(time.Millisecond)
					pointOnePercent := onePercent / 10
					if pointOnePercent > 0 {
						topPointOneTasks := topK(top1pTasks, pointOnePercent)
						task99Point9th := topPointOneTasks[0]
						report.Top99p9th = float32(task99Point9th.Duration) / float32(time.Millisecond)
					} else {
						report.Top99p9th = getTopDuration(top1pTasks)
					}
				} else {
					mostDuration := getTopDuration(top10pTasks)
					report.Top99th = mostDuration
					report.Top99p9th = mostDuration
				}
			} else {
				mostDuration := getTopDuration(tasks)
				report.Top90th = mostDuration
				report.Top99th = mostDuration
				report.Top99p9th = mostDuration
			}
		} else {
			mostDuration := getTopDuration(tasks)
			report.Median = mostDuration
			report.Top90th = mostDuration
			report.Top99th = mostDuration
			report.Top99p9th = mostDuration
		}
	}

	log(report)
}

func (c *metricsContainer) RemoveAll() interface{} {
	tasks := c.tasks
	duration := c.duration
	c.tasks = nil
	c.duration = 0

	return tasksDurationPair{
		tasks:    tasks,
		duration: duration,
	}
}

func getTopDuration(tasks []Task) float32 {
	top := topK(tasks, 1)
	if len(top) < 1 {
		return 0
	} else {
		return float32(top[0].Duration) / float32(time.Millisecond)
	}
}

func log(report *Report) {
	writeReport(report)
	logx.Statf("(%s) - tps: %.1f/s, avg time: %.1fms, med: %1fms, 90th: %.1fms, 99th: %.1fms, 99.9th: %.1fms",
		report.Name, report.ReqsPerSecond, report.Average, report.Median, report.Top90th,
		report.Top99th, report.Top99p9th)
}

func writeReport(report *Report) {
	writerLock.Lock()
	defer writerLock.Unlock()

	if reportWriter != nil {
		if err := reportWriter.Write(report); err != nil {
			logx.Error(err)
		}
	}
}
