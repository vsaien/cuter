package httphandler

import (
	"net/http"
	"time"

	"github.com/vsaien/cuter/lib/traffic"
)

func TrafficHandler(metrics *traffic.Metrics) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			startTime := time.Now()
			defer func() {
				metrics.Add(traffic.Task{
					Duration: time.Since(startTime),
				})
			}()

			next.ServeHTTP(w, r)
		})
	}
}
