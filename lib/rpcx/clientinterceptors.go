package rpcx

import (
	"context"
	"path"
	"time"

	"github.com/vsaien/cuter/lib/breaker"
	"github.com/vsaien/cuter/lib/logx"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	defaultTimeout      = time.Second * 2
	clientSlowThreshold = time.Millisecond * 500
)

func acceptable(err error) bool {
	switch status.Code(err) {
	case codes.DeadlineExceeded, codes.Unimplemented, codes.Internal, codes.Unavailable, codes.DataLoss:
		return false
	default:
		return true
	}
}

func buildClientTimeoutInterceptor(timeout time.Duration) grpc.UnaryClientInterceptor {
	if timeout == 0 {
		timeout = defaultTimeout
	}

	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		ctx, cancel := context.WithDeadline(ctx, time.Now().Add(timeout))
		defer cancel()
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

func clientBreakerInterceptor(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn,
	invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	breakerName := path.Join(cc.Target(), method)
	return breaker.DoWithAcceptable(breakerName, func() error {
		return invoker(ctx, method, req, reply, cc, opts...)
	}, acceptable)
}

func clientDurationInterceptor(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn,
	invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	serverName := path.Join(cc.Target(), method)
	start := time.Now()
	err := invoker(ctx, method, req, reply, cc, opts...)
	if err != nil {
		logx.Infof("fail - %s - %s - %v - %s", time.Since(start), serverName, req, err.Error())
	} else {
		elapsed := time.Since(start)
		if elapsed > clientSlowThreshold {
			logx.Slowf("[RPC] ok - slowcall(%s) - %s - %v - %v", elapsed, serverName, req, reply)
		}
	}

	return err
}
