package redis

import (
	"fmt"

	"github.com/vsaien/cuter/lib/logx"

	red "github.com/go-redis/redis"
)

type RedisClosableNode interface {
	RedisNode
	Close()
}

func CreateRedisBlockingNode(r *Redis) (RedisClosableNode, error) {
	timeout := readWriteTimeout + blockingQueryTimeout

	switch r.RedisType {
	case NodeType:
		client := red.NewClient(&red.Options{
			Addr:         r.RedisAddr,
			Password:     r.RedisPass,
			DB:           defaultDatabase,
			MaxRetries:   maxRetries,
			PoolSize:     1,
			MinIdleConns: 1,
			ReadTimeout:  timeout,
		})
		return &redisClientBridge{client}, nil
	case ClusterType:
		client := red.NewClusterClient(&red.ClusterOptions{
			Addrs:        []string{r.RedisAddr},
			Password:     r.RedisPass,
			MaxRetries:   maxRetries,
			PoolSize:     1,
			MinIdleConns: 1,
			ReadTimeout:  timeout,
		})
		return &redisClusterBridge{client}, nil
	default:
		return nil, fmt.Errorf("unknown redis type: %s", r.RedisType)
	}
}

type (
	redisClientBridge struct {
		*red.Client
	}

	redisClusterBridge struct {
		*red.ClusterClient
	}
)

func (bridge *redisClientBridge) Close() {
	if err := bridge.Client.Close(); err != nil {
		logx.Errorf("Error occurred on close redis client: %s", err)
	}
}

func (bridge *redisClusterBridge) Close() {
	if err := bridge.ClusterClient.Close(); err != nil {
		logx.Errorf("Error occurred on close redis cluster: %s", err)
	}
}
