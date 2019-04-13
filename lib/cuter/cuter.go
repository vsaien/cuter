package cuter

import (
	"log"
	"net/http"

	"github.com/vsaien/cuter/lib/httprouter"
	"github.com/vsaien/cuter/lib/logx"
)

type (
	runOptions struct {
		start func(*server) error
	}

	RunOption func(*Engine)

	Engine struct {
		srv  *server
		opts runOptions
	}
)

func MustNewEngine(c ServerConfig, opts ...RunOption) *Engine {
	engine, err := NewEngine(c, opts...)
	if err != nil {
		log.Fatal(err)
	}

	return engine
}

func NewEngine(c ServerConfig, opts ...RunOption) (*Engine, error) {
	engine := &Engine{
		srv: newServer(c),
		opts: runOptions{
			start: startWithPatRouter,
		},
	}

	for _, opt := range opts {
		opt(engine)
	}

	if err := c.SetUp(); err != nil {
		return nil, err
	} else {
		return engine, nil
	}
}

func (e *Engine) AddRoutes(rs []Route, opts ...RouteOption) {
	r := featuredRoutes{
		routes: rs,
	}
	for _, opt := range opts {
		opt(&r)
	}
	e.srv.AddRoutes(r)
}

func (e *Engine) AddRoute(r Route, opts ...RouteOption) {
	e.AddRoutes([]Route{r}, opts...)
}

func (e *Engine) Start() {
	handleError(e.opts.start(e.srv))
}

func (e *Engine) Stop() {
	logx.Close()
}

func (e *Engine) Use(middleware Middleware) {
	e.srv.use(middleware)
}

func handleError(err error) {
	// ErrServerClosed means the server is closed manually
	if err == nil || err == http.ErrServerClosed {
		return
	}

	logx.Error(err)
	panic(err)
}

func startWithPatRouter(srv *server) error {
	return srv.StartWithRouter(httprouter.NewPatRouter())
}

func validateSecret(secret string) {
	if len(secret) < 8 {
		panic("secret's length can't be less than 8")
	}
}
