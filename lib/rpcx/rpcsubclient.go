package rpcx

import (
	"github.com/vsaien/cuter/lib/etcd"

	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
)

type (
	RoundRobinSubClient struct {
		*etcd.RoundRobinSubClient
	}

	ConsistentSubClient struct {
		*etcd.ConsistentSubClient
	}
)

func NewRoundRobinRpcClient(conf etcd.EtcdConf, opts ...ClientOption) (*RoundRobinSubClient, error) {
	options := buildDialOptions(opts...)
	subClient, err := etcd.NewRoundRobinSubClient(conf, func(server string) (interface{}, error) {
		return grpc.Dial(server, options...)
	}, func(server string, conn interface{}) error {
		return conn.(*grpc.ClientConn).Close()
	}, etcd.Exclusive())
	if err != nil {
		return nil, err
	} else {
		return &RoundRobinSubClient{subClient}, nil
	}
}

func NewConsistentRpcClient(conf etcd.EtcdConf, opts ...ClientOption) (*ConsistentSubClient, error) {
	options := buildDialOptions(opts...)
	subClient, err := etcd.NewConsistentSubClient(conf, func(server string) (interface{}, error) {
		return grpc.Dial(server, options...)
	}, func(server string, conn interface{}) error {
		return conn.(*grpc.ClientConn).Close()
	})
	if err != nil {
		return nil, err
	} else {
		return &ConsistentSubClient{subClient}, nil
	}
}

func (cli *RoundRobinSubClient) Next() (*grpc.ClientConn, bool) {
	for {
		v, ok := cli.RoundRobinSubClient.Next()
		if !ok {
			break
		}

		conn, yes := v.(*grpc.ClientConn)
		if !yes {
			break
		}

		if conn.GetState() == connectivity.Ready {
			return conn, true
		}
	}

	return nil, false
}

func (cli *ConsistentSubClient) Next(key string) (*grpc.ClientConn, bool) {
	if v, ok := cli.ConsistentSubClient.Next(key); ok {
		conn, ok := v.(*grpc.ClientConn)
		return conn, ok
	}

	return nil, false
}
