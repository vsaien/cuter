package messages

import (
	"net/http"

	"github.com/vsaien/cuter/lib/httplog"
	"github.com/vsaien/cuter/lib/logx"
	"github.com/vsaien/cuter/lib/mapping"
)

type HttpMessage struct {
	Request    *http.Request
	Data       map[string]interface{}
	UserId     uint64
	DeviceUuid string
	Logger     logx.Logger
}

func NewHttpMessage(r *http.Request, data map[string]interface{}) *HttpMessage {
	return &HttpMessage{
		Request: r,
		Data:    data,
		Logger:  &httpLogger{},
	}
}

func (message *HttpMessage) Fill(v interface{}) error {
	return mapping.UnmarshalKey(message.Data, v)
}

func (message *HttpMessage) SetAttr(key string, value interface{}) {
	message.Data[key] = value
}

type httpLogger struct {
	request *http.Request
}

func (hl *httpLogger) Error(v ...interface{}) {
	httplog.Error(hl.request, v...)
}

func (hl *httpLogger) Errorf(format string, v ...interface{}) {
	httplog.Errorf(hl.request, format, v...)
}

func (hl *httpLogger) Info(v ...interface{}) {
	httplog.Info(hl.request, v...)
}

func (hl *httpLogger) Infof(format string, v ...interface{}) {
	httplog.Infof(hl.request, format, v...)
}
