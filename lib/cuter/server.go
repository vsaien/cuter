package cuter

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/vsaien/cuter/lib/httphandler"
	"github.com/vsaien/cuter/lib/httprouter"
	"github.com/vsaien/cuter/lib/httpserver"
	"github.com/vsaien/cuter/lib/traffic"

	"github.com/justinas/alice"
)

var ErrSignatureConfig = errors.New("bad config for Signature")

type (
	Middleware func(handle http.HandlerFunc) http.HandlerFunc

	server struct {
		conf        ServerConfig
		routes      []featuredRoutes
		middlewares []Middleware
	}
)

func newServer(c ServerConfig) *server {
	return &server{
		conf: c,
	}
}

func (s *server) AddRoutes(r featuredRoutes) {
	s.routes = append(s.routes, r)
}

func (s *server) Start() error {
	return s.StartWithRouter(httprouter.NewPatRouter())
}

func (s *server) StartWithRouter(router httprouter.Router) error {
	return httpserver.StartHttp(s.conf.Host, s.conf.Port, router)
}

func (s *server) bindRoute(router httprouter.Router, metrics *traffic.Metrics, route Route,
	customize func(chain alice.Chain) alice.Chain) error {
	chain := alice.New(s.getLogHandler())
	chain = chain.Append(httphandler.MaxConns(s.conf.MaxConns),
		httphandler.TimeoutHandler(time.Duration(s.conf.Timeout)*time.Millisecond),
		httphandler.RecoverHandler,
		httphandler.TrafficHandler(metrics))
	chain = customize(chain)

	for _, middleware := range s.middlewares {
		chain = chain.Append(convertMiddleware(middleware))
	}
	handle := chain.ThenFunc(route.Handler)

	return router.Handle(route.Method, route.Path, handle)
}

func (s *server) createMetrics() *traffic.Metrics {
	var metrics *traffic.Metrics

	if len(s.conf.Name) > 0 {
		metrics = traffic.NewMetrics(s.conf.Name)
	} else {
		metrics = traffic.NewMetrics(fmt.Sprintf("%s:%d", s.conf.Host, s.conf.Port))
	}

	return metrics
}

func (s *server) getLogHandler() func(http.Handler) http.Handler {
	if s.conf.Verbose {
		return httphandler.DetailedLogHandler
	} else {
		return httphandler.LogHandler
	}
}

func (s *server) use(middleware Middleware) {
	s.middlewares = append(s.middlewares, middleware)
}

func convertMiddleware(ware Middleware) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(ware(next.ServeHTTP))
	}
}
