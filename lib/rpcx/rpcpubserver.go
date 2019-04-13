package rpcx

import (
	"github.com/vsaien/cuter/lib/etcd"

	"google.golang.org/grpc"
)

// etcdEndpoints []string, etcdKey, listenOn, userName, password string
func NewRpcPubServer(c RpcServerConf, opts ...grpc.ServerOption) (Server, error) {
	registerEtcd := func() error {
		pubClient := etcd.NewPublisher(c.Etcd.Hosts, c.Etcd.Key, c.ListenOn, c.Etcd.UserName, c.Etcd.Password)
		return pubClient.KeepAlive()
	}
	server := keepAliveServer{
		registerEtcd: registerEtcd,
		Server:       NewRpcServer(c.ListenOn),
	}
	server.AddOptions(opts...)

	return server, nil
}

type keepAliveServer struct {
	registerEtcd func() error
	Server
}

func (ags keepAliveServer) Start(fn RegisterFn) error {
	if err := ags.registerEtcd(); err != nil {
		return err
	}

	return ags.Server.Start(fn)
}
