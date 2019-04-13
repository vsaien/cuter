package rpcx

import (
	"log"
	"time"

	"github.com/vsaien/cuter/lib/logx"
	"github.com/vsaien/cuter/lib/traffic"
	"google.golang.org/grpc"
)

type (
	RegisterFn func(*grpc.Server)

	Server interface {
		AddOptions(options ...grpc.ServerOption)
		AddStreamInterceptors(interceptors ...grpc.StreamServerInterceptor)
		AddUnaryInterceptors(interceptors ...grpc.UnaryServerInterceptor)
		SetName(string)
		Start(register RegisterFn) error
	}

	baseRpcServer struct {
		address            string
		metrics            *traffic.Metrics
		options            []grpc.ServerOption
		streamInterceptors []grpc.StreamServerInterceptor
		unaryInterceptors  []grpc.UnaryServerInterceptor
	}
	RpcServer struct {
		server   Server
		register RegisterFn
	}
)

func newBaseRpcServer(address string) *baseRpcServer {
	metrics := traffic.NewMetrics(address)

	return &baseRpcServer{
		address: address,
		metrics: metrics,
	}
}

func (s *baseRpcServer) AddOptions(options ...grpc.ServerOption) {
	s.options = append(s.options, options...)
}

func (s *baseRpcServer) AddStreamInterceptors(interceptors ...grpc.StreamServerInterceptor) {
	s.streamInterceptors = append(s.streamInterceptors, interceptors...)
}

func (s *baseRpcServer) AddUnaryInterceptors(interceptors ...grpc.UnaryServerInterceptor) {
	s.unaryInterceptors = append(s.unaryInterceptors, interceptors...)
}

func (s *baseRpcServer) SetName(name string) {
	s.metrics.SetName(name)
}

func MustNewServer(c RpcServerConf, register RegisterFn) *RpcServer {
	server, err := NewServer(c, register)
	if err != nil {
		log.Fatal(err)
	}

	return server
}

func NewServer(c RpcServerConf, register RegisterFn) (*RpcServer, error) {
	var err error
	var server Server
	if c.HasEtcd() {
		server, err = NewRpcPubServer(c)
		if err != nil {
			return nil, err
		}
	} else {
		server = NewRpcServer(c.ListenOn)
	}

	server.SetName(c.Name)
	if err = setupInterceptors(server, c); err != nil {
		return nil, err
	}

	rpcServer := &RpcServer{
		server:   server,
		register: register,
	}
	if err = c.SetUp(); err != nil {
		return nil, err
	}

	return rpcServer, nil
}

func (rs *RpcServer) Start() {
	if err := rs.server.Start(rs.register); err != nil {
		logx.Error(err)
		panic(err)
	}
}

func (rs *RpcServer) Stop() {
	logx.Close()
}

func setupInterceptors(server Server, c RpcServerConf) error {
	if c.Timeout > 0 {
		server.AddUnaryInterceptors(UnaryTimeoutInterceptor(time.Duration(c.Timeout) * time.Millisecond))
	}
	return nil
}
