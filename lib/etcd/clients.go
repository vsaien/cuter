package etcd

import (
	"fmt"
	"strings"

	"go.etcd.io/etcd/clientv3"
)

const (
	indexOfKey = 0
	indexOfId  = 1
)

var (
	DialTimeout    = dialTimeout
	RequestTimeout = requestTimeout
	TimeToLive     = timeToLive
)

type taskFn func(*clientv3.Client) error

func execute(config clientv3.Config, task taskFn) error {
	cli, err := clientv3.New(config)
	if err != nil {
		return err
	}
	defer cli.Close()

	return task(cli)
}

func extract(etcdKey string, index int) (string, bool) {
	if index < 0 {
		return "", false
	}

	fields := strings.FieldsFunc(etcdKey, func(ch rune) bool {
		return ch == delimiter
	})
	if index >= len(fields) {
		return "", false
	}

	return fields[index], true
}

func extractId(etcdKey string) (string, bool) {
	return extract(etcdKey, indexOfId)
}

func extractKey(etcdKey string) (string, bool) {
	return extract(etcdKey, indexOfKey)
}

func makeEtcdKey(key string, id int64) string {
	return fmt.Sprintf("%s%c%d", key, delimiter, id)
}
