package rpcserver

import (
	"context"
	"github.com/vsaien/cuter/application/test/demo/logic"
	"github.com/vsaien/cuter/application/test/demo/rpcproto"
)

type DemoServer struct {
	demoLogic *logic.DemoLogic
}

func NewDemoServer(logic *logic.DemoLogic) *DemoServer {
	return &DemoServer{
		demoLogic: logic,
	}
}

func (s *DemoServer) DemoFuc(_ context.Context, r *rpcproto.DemoRequest) (*rpcproto.DemoResponse, error) {
	st, err := s.demoLogic.DemoTwo(&logic.DemoRequest{Token: r.Token})
	if err != nil {
		return nil, err
	}
	return &rpcproto.DemoResponse{Token: st}, nil
}
