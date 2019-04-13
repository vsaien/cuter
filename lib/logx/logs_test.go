package logx

import (
	"fmt"
	"io/ioutil"
	"log"
	"runtime"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var (
	s    = []byte("Sending #11 notification (id: 1451875113812010473) in #1 connection")
	pool = make(chan []byte, 1)
)

type mockWriter struct {
	builder strings.Builder
}

func (mw *mockWriter) Write(data []byte) (int, error) {
	return mw.builder.Write(data)
}

func (mw *mockWriter) Close() error {
	return nil
}

func (mw *mockWriter) Reset() {
	mw.builder.Reset()
}

func (mw *mockWriter) Contains(text string) bool {
	return strings.Index(mw.builder.String(), text) > -1
}

func TestFileLineFileMode(t *testing.T) {
	writer := new(mockWriter)
	errorLog = writer
	atomic.StoreUint32(&initialized, 1)
	file, line := getFileLine()
	Error("anything")
	assert.True(t, writer.Contains(fmt.Sprintf("%s:%d", file, line+1)))

	writer.Reset()
	file, line = getFileLine()
	Errorf("anything %s", "format")
	assert.True(t, writer.Contains(fmt.Sprintf("%s:%d", file, line+1)))
}

func TestFileLineConsoleMode(t *testing.T) {
	writer := new(mockWriter)
	writeConsole = true
	errorLog = newLogWriter(log.New(writer, "[ERROR] ", flags))
	atomic.StoreUint32(&initialized, 1)
	file, line := getFileLine()
	Error("anything")
	assert.True(t, writer.Contains(fmt.Sprintf("%s:%d", file, line+1)))

	writer.Reset()
	file, line = getFileLine()
	Errorf("anything %s", "format")
	assert.True(t, writer.Contains(fmt.Sprintf("%s:%d", file, line+1)))
}

func BenchmarkCopyByteSliceAppend(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var buf []byte
		buf = append(buf, time.Now().Format(TimeFormat)...)
		buf = append(buf, ' ')
		buf = append(buf, s...)
		_ = buf
	}
}

func BenchmarkCopyByteSliceAllocExactly(b *testing.B) {
	for i := 0; i < b.N; i++ {
		now := []byte(time.Now().Format(TimeFormat))
		buf := make([]byte, len(now)+1+len(s))
		n := copy(buf, now)
		buf[n] = ' '
		copy(buf[n+1:], s)
	}
}

func BenchmarkCopyByteSlice(b *testing.B) {
	var buf []byte
	for i := 0; i < b.N; i++ {
		buf = make([]byte, len(s))
		copy(buf, s)
	}
	fmt.Fprint(ioutil.Discard, buf)
}

func BenchmarkCopyOnWriteByteSlice(b *testing.B) {
	var buf []byte
	for i := 0; i < b.N; i++ {
		size := len(s)
		buf = s[:size:size]
	}
	fmt.Fprint(ioutil.Discard, buf)
}

func BenchmarkCacheByteSlice(b *testing.B) {
	for i := 0; i < b.N; i++ {
		dup := fetch()
		copy(dup, s)
		put(dup)
	}
}

func BenchmarkLogs(b *testing.B) {
	b.ReportAllocs()

	log.SetOutput(ioutil.Discard)
	for i := 0; i < b.N; i++ {
		Info(i)
	}
}

func fetch() []byte {
	select {
	case b := <-pool:
		return b
	default:
	}
	return make([]byte, 4096)
}

func getFileLine() (string, int) {
	_, file, line, _ := runtime.Caller(1)
	short := file

	for i := len(file) - 1; i > 0; i-- {
		if file[i] == '/' {
			short = file[i+1:]
			break
		}
	}

	return short, line
}

func put(b []byte) {
	select {
	case pool <- b:
	default:
	}
}
