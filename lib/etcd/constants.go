package etcd

import "time"

const (
	dialTimeout          = 5 * time.Second
	renewDuration        = 4 * time.Second
	requestTimeout       = 3 * time.Second
	timeToLive     int64 = 10
	delimiter            = '/'
)
