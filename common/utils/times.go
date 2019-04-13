package utils

import (
	"fmt"
	"time"
)

type ElapsedTimer struct {
	start time.Time
}

func NewElapsedTimer() *ElapsedTimer {
	return &ElapsedTimer{
		start: time.Now(),
	}
}

func (et *ElapsedTimer) Duration() time.Duration {
	return time.Since(et.start)
}

func (et *ElapsedTimer) Elapsed() string {
	return fmt.Sprintf("%v", time.Since(et.start))
}

func (et *ElapsedTimer) ElapsedMs() string {
	return fmt.Sprintf("%.1fms", float32(time.Since(et.start))/float32(time.Millisecond))
}

func CurrentMicros() int64 {
	return time.Now().UnixNano() / int64(time.Microsecond)
}

func CurrentMillis() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}
