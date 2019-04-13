package etcd

import (
	"context"

	"github.com/vsaien/cuter/lib/logx"
	"github.com/vsaien/cuter/lib/threading"

	"go.etcd.io/etcd/clientv3"
)

type monitor struct {
	conf           EtcdConf
	fullCallback   fullCallbackFn
	changeCallback changeCallbackFn
}

func newMonitor(conf EtcdConf, fullCallback fullCallbackFn,
	changeCallback changeCallbackFn) *monitor {
	return &monitor{
		conf:           conf,
		fullCallback:   fullCallback,
		changeCallback: changeCallback,
	}
}

func (m *monitor) load() error {
	var kvs []keyValue
	if err := execute(clientv3.Config{
		Endpoints:   m.conf.Hosts,
		Username:    m.conf.UserName,
		Password:    m.conf.Password,
		DialTimeout: DialTimeout,
	}, func(client *clientv3.Client) error {
		ctx, cancel := context.WithTimeout(client.Ctx(), RequestTimeout)
		resp, err := client.Get(ctx, m.conf.Key, clientv3.WithPrefix())
		cancel()
		if err != nil {
			return err
		}

		for _, ev := range resp.Kvs {
			kvs = append(kvs, keyValue{
				key:   string(ev.Key),
				value: string(ev.Value),
			})
		}

		return nil
	}); err != nil {
		return err
	}

	m.fullCallback(kvs)
	return nil
}

func (m *monitor) match(etcdKey string) bool {
	if key, ok := extractKey(etcdKey); ok {
		return key == m.conf.Key
	}

	return false
}

func (m *monitor) watch() error {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   m.conf.Hosts,
		DialTimeout: DialTimeout,
		Username:    m.conf.UserName,
		Password:    m.conf.Password,
	})
	if err != nil {
		return err
	}

	threading.GoSafe(func() {
		defer cli.Close()

		rch := cli.Watch(cli.Ctx(), m.conf.Key, clientv3.WithPrefix())
		for wresp := range rch {
			for _, ev := range wresp.Events {
				switch ev.Type {
				case clientv3.EventTypePut:
					m.changeCallback(ADD, keyValue{
						key:   string(ev.Kv.Key),
						value: string(ev.Kv.Value),
					})
				case clientv3.EventTypeDelete:
					m.changeCallback(DELETE, keyValue{
						key:   string(ev.Kv.Key),
						value: string(ev.Kv.Value),
					})
				default:
					logx.Errorf("Unknown event type: %v", ev.Type)
				}
			}
		}
	})

	return nil
}
