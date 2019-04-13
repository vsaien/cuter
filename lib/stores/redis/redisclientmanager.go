package redis

import (
	"sync"

	"github.com/vsaien/cuter/lib/syncx"

	red "github.com/go-redis/redis"
)

const (
	defaultDatabase      = 0
	maxRetries           = 3
	idleRedisConnections = 8
)

var redisClientManager *RedisClientManager

type RedisClientManager struct {
	stores         map[string]*red.Client
	exclusiveCalls syncx.ExclusiveCalls
	lock           *sync.Mutex
}

func init() {
	redisClientManager = &RedisClientManager{
		stores:         make(map[string]*red.Client),
		exclusiveCalls: syncx.NewExclusiveCalls(),
		lock:           new(sync.Mutex),
	}
}

func GetRedisClient(server, pass string) (*red.Client, error) {
	val, err := redisClientManager.exclusiveCalls.Do(server, func() (interface{}, error) {
		redisClientManager.lock.Lock()
		store, ok := redisClientManager.stores[server]
		redisClientManager.lock.Unlock()
		if ok {
			return store, nil
		} else {
			store = red.NewClient(&red.Options{
				Addr:         server,
				Password:     pass,
				DB:           defaultDatabase,
				MaxRetries:   maxRetries,
				MinIdleConns: idleRedisConnections,
			})
			store.WrapProcess(process)

			redisClientManager.lock.Lock()
			redisClientManager.stores[server] = store
			redisClientManager.lock.Unlock()

			return store, nil
		}
	})
	if err != nil {
		return nil, err
	}

	return val.(*red.Client), nil
}
