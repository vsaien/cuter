package redis

import (
	"sync"

	"github.com/vsaien/cuter/lib/syncx"

	red "github.com/go-redis/redis"
)

var redisClusterManager *RedisClusterManager

type RedisClusterManager struct {
	stores         map[string]*red.ClusterClient
	exclusiveCalls syncx.ExclusiveCalls
	lock           *sync.Mutex
}

func init() {
	redisClusterManager = &RedisClusterManager{
		stores:         make(map[string]*red.ClusterClient),
		exclusiveCalls: syncx.NewExclusiveCalls(),
		lock:           new(sync.Mutex),
	}
}

func GetRedisCluster(server, pass string) (*red.ClusterClient, error) {
	val, err := redisClusterManager.exclusiveCalls.Do(server, func() (interface{}, error) {
		redisClusterManager.lock.Lock()
		store, ok := redisClusterManager.stores[server]
		redisClusterManager.lock.Unlock()
		if ok {
			return store, nil
		} else {
			store = red.NewClusterClient(&red.ClusterOptions{
				Addrs:        []string{server},
				Password:     pass,
				MaxRetries:   maxRetries,
				MinIdleConns: idleRedisConnections,
			})
			store.WrapProcess(process)

			redisClusterManager.lock.Lock()
			redisClusterManager.stores[server] = store
			redisClusterManager.lock.Unlock()

			return store, nil
		}
	})
	if err != nil {
		return nil, err
	}

	return val.(*red.ClusterClient), nil
}
