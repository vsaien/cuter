package httplog

import (
	"bytes"
	"fmt"
	"net/http"
	"sync"

	"github.com/vsaien/cuter/lib/httpx"
	"github.com/vsaien/cuter/lib/logx"
)

const LogContext = "request_logs"

type LogCollector struct {
	Messages []string
	lock     sync.Mutex
}

func (lc *LogCollector) Append(msg string) {
	timedMsg := logx.AddTime(msg)
	lc.lock.Lock()
	lc.Messages = append(lc.Messages, timedMsg)
	lc.lock.Unlock()
}

func (lc *LogCollector) Flush() string {
	var buffer bytes.Buffer

	start := true
	for _, message := range lc.takeAll() {
		if start {
			start = false
		} else {
			buffer.WriteByte('\n')
		}
		buffer.WriteString(message)
	}

	return buffer.String()
}

func (lc *LogCollector) takeAll() []string {
	lc.lock.Lock()
	messages := lc.Messages
	lc.Messages = nil
	lc.lock.Unlock()

	return messages
}

func AppendLog(r *http.Request, message string) {
	logs := r.Context().Value(LogContext)
	if logs != nil {
		logs.(*LogCollector).Append(message)
	}
}

func Error(r *http.Request, v ...interface{}) {
	logx.ErrorCaller(1, Format(r, v...))
}

func Errorf(r *http.Request, format string, v ...interface{}) {
	logx.ErrorCaller(1, Formatf(r, format, v...))
}

func Format(r *http.Request, v ...interface{}) string {
	return formatWithReq(r, fmt.Sprint(v...))
}

func Formatf(r *http.Request, format string, v ...interface{}) string {
	return formatWithReq(r, fmt.Sprintf(format, v...))
}

func Info(r *http.Request, v ...interface{}) {
	AppendLog(r, Format(r, v...))
}

func Infof(r *http.Request, format string, v ...interface{}) {
	AppendLog(r, Formatf(r, format, v...))
}

func formatWithReq(r *http.Request, v string) string {
	return fmt.Sprintf("(%s - %s) %s", r.RequestURI, httpx.GetRemoteAddr(r), v)
}
