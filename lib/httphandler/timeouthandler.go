package httphandler

import (
	"net/http"
	"time"
)

const reason = "Request Timeout"

func TimeoutHandler(duration time.Duration) func(http.Handler) http.Handler {
	return func(handler http.Handler) http.Handler {
		if duration > 0 {
			return http.TimeoutHandler(handler, duration, reason)
		} else {
			return handler
		}
	}
}
