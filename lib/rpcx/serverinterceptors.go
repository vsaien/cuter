package rpcx

import (
	"context"
	"encoding/json"
	"time"

	"github.com/vsaien/cuter/lib/logx"
	"github.com/vsaien/cuter/lib/traffic"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"runtime/debug"
)

const serverSlowThreshold = time.Millisecond * 500

func StreamStatInterceptor(metrics *traffic.Metrics) grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo,
		handler grpc.StreamHandler) (err error) {
		defer handleCrash(func(r interface{}) {
			err = toPanicError(r)
		})

		return handler(srv, stream)
	}
}

func UnaryStatInterceptor(metrics *traffic.Metrics) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler) (resp interface{}, err error) {
		defer handleCrash(func(r interface{}) {
			err = toPanicError(r)
		})

		startTime := time.Now()
		defer func() {
			duration := time.Since(startTime)
			metrics.Add(traffic.Task{
				Duration: duration,
			})
			logDuration(info.FullMethod, req, duration)
		}()

		return handler(ctx, req)
	}
}

func UnaryTimeoutInterceptor(timeout time.Duration) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler) (resp interface{}, err error) {
		ctx, cancel := context.WithDeadline(ctx, time.Now().Add(timeout))
		defer cancel()
		return handler(ctx, req)
	}
}

func handleCrash(handler func(interface{})) {
	if r := recover(); r != nil {
		handler(r)
	}
}

func logDuration(method string, req interface{}, duration time.Duration) {
	content, err := json.Marshal(req)
	if err != nil {
		logx.Error(err)
	} else if duration > serverSlowThreshold {
		logx.Slowf("[RPC] slowcall(%s) - %s - %s", duration, method, string(content))
	} else {
		logx.Infof("%s - %s - %s", duration, method, string(content))
	}
}

func toPanicError(r interface{}) error {
	logx.Errorf("%+v %s", r, debug.Stack())
	return status.Errorf(codes.Internal, "panic: %v", r)
}
