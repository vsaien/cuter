package rpcx

import (
	"github.com/vsaien/cuter/lib/etcd"
	"github.com/vsaien/cuter/lib/service"
)

type (
	RpcServerConf struct {
		service.ServiceConf
		ListenOn      string
		Etcd          etcd.EtcdConf `json:",optional"`
		StrictControl bool          `json:",optional"`
		Timeout       int64         `json:",optional"`
	}

	RpcClientConf struct {
		etcd.EtcdConf `json:",optional"`
		Server        string `json:",optional"`
		BlockDial     bool   `json:",default=false"`
		Timeout       int64  `json:",optional"`
	}
)

func NewEtcdClientConf(hosts []string, key, userName, password string) RpcClientConf {
	return RpcClientConf{
		EtcdConf: etcd.EtcdConf{
			Hosts:    hosts,
			Key:      key,
			UserName: userName,
			Password: password,
		},
	}
}

func (sc RpcServerConf) HasEtcd() bool {
	return len(sc.Etcd.Hosts) > 0 && len(sc.Etcd.Key) > 0
}
