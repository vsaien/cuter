package baseresponse

import (
	"errors"
	"net/http"

	"github.com/vsaien/cuter/common/baseerror"
	"github.com/vsaien/cuter/lib/logx"

	"github.com/vsaien/cuter/lib/httpx"
)

const (
	StatusOk         = 0
	StatusParamError = 1
)

type response struct {
	Code        int         `json:"code"`
	Description string      `json:"desc,omitempty"`
	Data        interface{} `json:"data,omitempty"`
}

func httpError(w http.ResponseWriter, httpCode, appCode int, err error) {
	var description string
	if err != nil {
		description = err.Error()
	}

	httpx.WriteJson(w, httpCode, response{
		Code:        appCode,
		Description: description,
	})
}

func httpOk(w http.ResponseWriter, payload interface{}) {
	respond(w, http.StatusOK, StatusOk, payload)
}

func httpParamError(w http.ResponseWriter, reason string) {
	httpx.WriteJson(w, http.StatusBadRequest, response{
		Code:        StatusParamError,
		Description: reason,
	})
}

func respond(w http.ResponseWriter, httpCode, appCode int, data interface{}) {
	httpx.WriteJson(w, httpCode, response{
		Code: appCode,
		Data: data,
	})
}
func FormatResponse(data interface{}, err error, w http.ResponseWriter) {
	if err != nil {
		codeErr, ok := baseerror.FromError(err)
		if ok {
			httpError(w, http.StatusNotAcceptable, codeErr.Code(), codeErr)
		} else {
			httpError(w, http.StatusInternalServerError, -1, errors.New("访问出错!请一会儿再试试吧"))
		}
		logx.Error(err)
	} else {
		httpOk(w, data)
	}
}

func HttpParamError(w http.ResponseWriter, err error) {
	logx.Error(err)
	httpParamError(w, err.Error())
}

func HttpError(w http.ResponseWriter, httpCode, appCode int, err error) {
	logx.Error(err)
	httpError(w, httpCode, appCode, err)
}
