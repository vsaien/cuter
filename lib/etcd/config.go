package etcd

import "errors"

type EtcdConf struct {
	Hosts    []string
	Key      string
	UserName string
	Password string
}

func (c EtcdConf) Validate() error {
	if len(c.Hosts) == 0 {
		return errors.New("empty etcd hosts")
	} else if len(c.Key) == 0 {
		return errors.New("empty etcd key")
	} else if len(c.UserName) == 0 {
		return errors.New("empty etcd UserName")
	} else if len(c.Password) == 0 {
		return errors.New("empty etcd Password")
	} else {
		return nil
	}
}
