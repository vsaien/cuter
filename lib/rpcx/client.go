package rpcx

import (
	"log"
	"time"

	"github.com/vsaien/cuter/lib/etcd"
	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer/roundrobin"
	"google.golang.org/grpc/connectivity"
)

type (
	ClientOptions struct {
		Timeout     time.Duration
		DialOptions []grpc.DialOption
	}

	ClientOption func(options *ClientOptions)

	Client interface {
		Next() (*grpc.ClientConn, bool)
	}
	DirectClient struct {
		conn *grpc.ClientConn
	}
	RpcClient struct {
		client Client
	}
)

func WithDialOption(opt grpc.DialOption) ClientOption {
	return func(options *ClientOptions) {
		options.DialOptions = append(options.DialOptions, opt)
	}
}

func WithTimeout(timeout time.Duration) ClientOption {
	return func(options *ClientOptions) {
		options.Timeout = timeout
	}
}

func buildDialOptions(opts ...ClientOption) []grpc.DialOption {
	var clientOptions ClientOptions
	for _, opt := range opts {
		opt(&clientOptions)
	}

	timeoutInterceptor := buildClientTimeoutInterceptor(clientOptions.Timeout)
	options := []grpc.DialOption{
		grpc.WithInsecure(),
		WithUnaryClientInterceptors(
			clientBreakerInterceptor,
			clientDurationInterceptor,
			timeoutInterceptor,
		),
	}

	return append(options, clientOptions.DialOptions...)
}

func NewDirectClient(server string, opts ...ClientOption) (*DirectClient, error) {
	opts = append(opts, WithDialOption(grpc.WithBalancerName(roundrobin.Name)))
	options := buildDialOptions(opts...)
	conn, err := grpc.Dial(server, options...)
	if err != nil {
		return nil, err
	}

	return &DirectClient{
		conn: conn,
	}, nil
}

func (c *DirectClient) Next() (*grpc.ClientConn, bool) {
	state := c.conn.GetState()
	if state == connectivity.Ready {
		return c.conn, true
	} else {
		return nil, false
	}
}

func MustNewClient(c RpcClientConf) *RpcClient {
	cli, err := NewClient(c)
	if err != nil {
		log.Fatalf("rpc client %+v failed: %s", c, err.Error())
	}

	return cli
}

func NewClient(c RpcClientConf) (*RpcClient, error) {
	opts := []ClientOption{}
	if c.BlockDial {
		opts = append(opts, WithDialOption(grpc.WithBlock()))
	}
	if c.Timeout > 0 {
		opts = append(opts, WithTimeout(time.Duration(c.Timeout)*time.Millisecond))
	}

	var client Client
	var err error
	if len(c.Server) > 0 {
		client, err = NewDirectClient(c.Server, opts...)
	} else if err = c.EtcdConf.Validate(); err == nil {
		client, err = NewRoundRobinRpcClient(c.EtcdConf, opts...)
	}
	if err != nil {
		return nil, err
	}

	return &RpcClient{
		client: client,
	}, nil
}

func NewClientNoAuth(c etcd.EtcdConf) (*RpcClient, error) {
	client, err := NewRoundRobinRpcClient(c)
	if err != nil {
		return nil, err
	}

	return &RpcClient{
		client: client,
	}, nil
}

func (rc *RpcClient) Next() (*grpc.ClientConn, bool) {
	return rc.client.Next()
}
