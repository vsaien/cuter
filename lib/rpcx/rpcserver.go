package rpcx

import (
	"net"

	"github.com/vsaien/cuter/lib/system"

	"google.golang.org/grpc"
)

type rpcServer struct {
	*baseRpcServer
}

func init() {
	InitLogger()
}

func NewRpcServer(address string) Server {
	return &rpcServer{
		baseRpcServer: newBaseRpcServer(address),
	}
}

func (s *rpcServer) SetName(name string) {
	s.baseRpcServer.SetName(name)
}

func (s *rpcServer) Start(register RegisterFn) error {
	lis, err := net.Listen("tcp", s.address)
	if err != nil {
		return err
	}

	unaryInterceptors := []grpc.UnaryServerInterceptor{UnaryStatInterceptor(s.metrics)}
	unaryInterceptors = append(unaryInterceptors, s.unaryInterceptors...)
	streamInterceptors := []grpc.StreamServerInterceptor{StreamStatInterceptor(s.metrics)}
	streamInterceptors = append(streamInterceptors, s.streamInterceptors...)
	options := append(s.options, WithUnaryServerInterceptors(unaryInterceptors...),
		WithStreamServerInterceptors(streamInterceptors...))
	server := grpc.NewServer(options...)
	register(server)
	// we need to make sure all others are wrapped up
	// so we do graceful stop at shutdown phase instead of wrap up phase
	shutdownCalled := system.AddShutdownListener(func() {
		server.GracefulStop()
	})
	err = server.Serve(lis)
	shutdownCalled()

	return err
}
