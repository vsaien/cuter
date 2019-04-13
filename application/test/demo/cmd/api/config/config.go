package config

import (
	"github.com/vsaien/cuter/lib/cuter"
	"github.com/vsaien/cuter/lib/rpcx"
	"github.com/vsaien/cuter/lib/stores/redis"
)

type (
	Config struct {
		cuter.ServerConfig
		Mysql struct {
			DataSource string
			Table      struct {
				Demo string
			}
		}
		Mongodb struct {
			Url        string
			Database   string
			Collection struct {
				Demo string
			}
			Concurrency int `json:",default=100"`
			Timeout     int `json:",default=1000"`
		}
		BizRedis   redis.RedisConf
		CacheRedis redis.RedisConf
		DemoRpc    rpcx.RpcClientConf
	}
)
