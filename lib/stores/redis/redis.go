package redis

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	red "github.com/go-redis/redis"
)

const (
	ClusterType = "cluster"
	NodeType    = "node"
	Nil         = red.Nil

	blockingQueryTimeout = 5 * time.Second
	readWriteTimeout     = 2 * time.Second

	slowThreshold = time.Millisecond * 100
)

var (
	ErrNilNode  = errors.New("nil redis node")
	ErrNotFound = errors.New("key not found")
	ErrTimeout  = errors.New("timeout on talkin to redis")
)

type (
	Pair struct {
		Key   string
		Score int64
	}

	// thread-safe
	Redis struct {
		RedisAddr string
		RedisType string
		RedisPass string
	}

	RedisNode interface {
		red.Cmdable
	}
)

func NewRedis(redisAddr, redisType string, redisPass ...string) *Redis {
	var pass string
	for _, v := range redisPass {
		pass = v
	}

	return &Redis{
		RedisAddr: redisAddr,
		RedisType: redisType,
		RedisPass: pass,
	}
}

// Use passed in redis connection to execute blocking queries
// Doesn't benefit from pooling redis connections of blocking queries
func (s *Redis) Blpop(redisNode RedisNode, key string) (string, error) {
	if redisNode == nil {
		return "", ErrNilNode
	}

	vals, err := redisNode.BLPop(blockingQueryTimeout, key).Result()
	if err != nil {
		return "", err
	}

	if len(vals) < 2 {
		return "", fmt.Errorf("no value on key: %s", key)
	} else {
		return vals[1], nil
	}
}

func (s *Redis) BlpopEx(redisNode RedisNode, key string) (string, bool, error) {
	if redisNode == nil {
		return "", false, ErrNilNode
	}

	vals, err := redisNode.BLPop(blockingQueryTimeout, key).Result()
	if err != nil {
		return "", false, err
	}

	if len(vals) < 2 {
		return "", false, fmt.Errorf("no value on key: %s", key)
	} else {
		return vals[1], true, nil
	}
}

func (s *Redis) Del(keys ...string) (int, error) {
	conn, err := getRedis(s)
	if err != nil {
		return 0, err
	}

	if val, err := conn.Del(keys...).Result(); err != nil {
		return 0, err
	} else {
		return int(val), nil
	}
}

func (s *Redis) Eval(script string, keys []string, args ...interface{}) (interface{}, error) {
	conn, err := getRedis(s)
	if err != nil {
		return nil, err
	}

	return conn.Eval(script, keys, args...).Result()
}

func (s *Redis) Exists(key string) (bool, error) {
	conn, err := getRedis(s)
	if err != nil {
		return false, err
	}

	if val, err := conn.Exists(key).Result(); err != nil {
		return false, err
	} else {
		return val == 1, nil
	}
}

func (s *Redis) Expire(key string, seconds int) error {
	conn, err := getRedis(s)
	if err != nil {
		return err
	}

	return conn.Expire(key, time.Duration(seconds)*time.Second).Err()
}

func (s *Redis) Expireat(key string, expireTime int64) error {
	conn, err := getRedis(s)
	if err != nil {
		return err
	}

	return conn.ExpireAt(key, time.Unix(expireTime, 0)).Err()
}

func (s *Redis) Get(key string) (string, error) {
	conn, err := getRedis(s)
	if err != nil {
		return "", err
	}

	if val, err := conn.Get(key).Result(); err == red.Nil {
		return "", nil
	} else if err != nil {
		return "", err
	} else {
		return val, nil
	}
}

func (s *Redis) Hdel(key, field string) (bool, error) {
	conn, err := getRedis(s)
	if err != nil {
		return false, err
	}

	if val, err := conn.HDel(key, field).Result(); err != nil {
		return false, err
	} else {
		return val == 1, nil
	}
}

func (s *Redis) Hexists(key, field string) (bool, error) {
	conn, err := getRedis(s)
	if err != nil {
		return false, err
	}

	return conn.HExists(key, field).Result()
}

func (s *Redis) Hget(key, field string) (string, error) {
	conn, err := getRedis(s)
	if err != nil {
		return "", err
	}

	return conn.HGet(key, field).Result()
}

func (s *Redis) Hgetall(key string) (map[string]string, error) {
	conn, err := getRedis(s)
	if err != nil {
		return nil, err
	}

	return conn.HGetAll(key).Result()
}

func (s *Redis) Hincrby(key, field string, increment int) (int, error) {
	conn, err := getRedis(s)
	if err != nil {
		return 0, err
	}

	if val, err := conn.HIncrBy(key, field, int64(increment)).Result(); err != nil {
		return 0, err
	} else {
		return int(val), nil
	}
}

func (s *Redis) Hkeys(key string) ([]string, error) {
	conn, err := getRedis(s)
	if err != nil {
		return nil, err
	}

	return conn.HKeys(key).Result()
}

func (s *Redis) Hlen(key string) (int, error) {
	conn, err := getRedis(s)
	if err != nil {
		return 0, err
	}

	if val, err := conn.HLen(key).Result(); err != nil {
		return 0, err
	} else {
		return int(val), nil
	}
}

func (s *Redis) Hmget(key string, fields ...string) ([]string, error) {
	conn, err := getRedis(s)
	if err != nil {
		return nil, err
	}

	if vals, err := conn.HMGet(key, fields...).Result(); err != nil {
		return nil, err
	} else {
		return toStrings(vals), nil
	}
}

func (s *Redis) Hset(key, field, value string) error {
	conn, err := getRedis(s)
	if err != nil {
		return err
	}

	return conn.HSet(key, field, value).Err()
}

func (s *Redis) Hsetnx(key, field, value string) (bool, error) {
	conn, err := getRedis(s)
	if err != nil {
		return false, err
	}

	return conn.HSetNX(key, field, value).Result()
}

func (s *Redis) Hmset(key string, fieldsAndValues map[string]string) error {
	conn, err := getRedis(s)
	if err != nil {
		return err
	}

	vals := make(map[string]interface{}, len(fieldsAndValues))
	for k, v := range fieldsAndValues {
		vals[k] = v
	}

	return conn.HMSet(key, vals).Err()
}

func (s *Redis) Hvals(key string) ([]string, error) {
	conn, err := getRedis(s)
	if err != nil {
		return nil, err
	}

	return conn.HVals(key).Result()
}

func (s *Redis) Incr(key string) (int64, error) {
	conn, err := getRedis(s)
	if err != nil {
		return 0, err
	}

	return conn.Incr(key).Result()
}

func (s *Redis) Incrby(key string, increment int64) (int64, error) {
	conn, err := getRedis(s)
	if err != nil {
		return 0, err
	}

	return conn.IncrBy(key, int64(increment)).Result()
}

func (s *Redis) Keys(pattern string) ([]string, error) {
	conn, err := getRedis(s)
	if err != nil {
		return nil, err
	}

	return conn.Keys(pattern).Result()
}

func (s *Redis) Llen(key string) (int, error) {
	conn, err := getRedis(s)
	if err != nil {
		return 0, err
	}

	if val, err := conn.LLen(key).Result(); err != nil {
		return 0, err
	} else {
		return int(val), nil
	}
}

func (s *Redis) Lpop(key string) (string, error) {
	conn, err := getRedis(s)
	if err != nil {
		return "", err
	}

	return conn.LPop(key).Result()
}

func (s *Redis) Lrange(key string, start int, stop int) ([]string, error) {
	conn, err := getRedis(s)
	if err != nil {
		return nil, err
	}

	return conn.LRange(key, int64(start), int64(stop)).Result()
}

func (s *Redis) Lrem(key string, count int, value string) (int, error) {
	conn, err := getRedis(s)
	if err != nil {
		return 0, err
	}

	if val, err := conn.LRem(key, int64(count), value).Result(); err != nil {
		return 0, err
	} else {
		return int(val), nil
	}
}

func (s *Redis) Mget(keys ...string) ([]string, error) {
	conn, err := getRedis(s)
	if err != nil {
		return nil, err
	}

	if vals, err := conn.MGet(keys...).Result(); err != nil {
		return nil, err
	} else {
		return toStrings(vals), nil
	}
}

func (s *Redis) Persist(key string) (bool, error) {
	conn, err := getRedis(s)
	if err != nil {
		return false, err
	}

	return conn.Persist(key).Result()
}

func (s *Redis) Pfadd(key string, values ...interface{}) (bool, error) {
	conn, err := getRedis(s)
	if err != nil {
		return false, err
	}

	if val, err := conn.PFAdd(key, values...).Result(); err != nil {
		return false, err
	} else {
		return val == 1, nil
	}
}

func (s *Redis) Pfcount(key string) (int64, error) {
	conn, err := getRedis(s)
	if err != nil {
		return 0, err
	}

	return conn.PFCount(key).Result()
}

func (s *Redis) Ping() bool {
	conn, err := getRedis(s)
	if err != nil {
		return false
	}

	if val, err := conn.Ping().Result(); err != nil {
		return false
	} else {
		return val == "PONG"
	}
}

func (s *Redis) Rpush(key string, values ...interface{}) (int, error) {
	conn, err := getRedis(s)
	if err != nil {
		return 0, err
	}

	if val, err := conn.RPush(key, values...).Result(); err != nil {
		return 0, err
	} else {
		return int(val), nil
	}
}

func (s *Redis) Lpush(key string, values ...interface{}) (int, error) {
	conn, err := getRedis(s)
	if err != nil {
		return 0, err
	}

	if val, err := conn.LPush(key, values...).Result(); err != nil {
		return 0, err
	} else {
		return int(val), nil
	}
}

func (s *Redis) Sadd(key string, values ...interface{}) (int, error) {
	conn, err := getRedis(s)
	if err != nil {
		return 0, err
	}

	if val, err := conn.SAdd(key, values...).Result(); err != nil {
		return 0, err
	} else {
		return int(val), nil
	}
}

func (s *Redis) Scan(cursor uint64, match string, count int64) ([]string, uint64, error) {
	conn, err := getRedis(s)
	if err != nil {
		return nil, 0, err
	}

	return conn.Scan(cursor, match, count).Result()
}

func (s *Redis) Scard(key string) (int64, error) {
	conn, err := getRedis(s)
	if err != nil {
		return 0, err
	}

	return conn.SCard(key).Result()
}

func (s *Redis) Set(key string, value string) error {
	conn, err := getRedis(s)
	if err != nil {
		return err
	}

	return conn.Set(key, value, 0).Err()
}

func (s *Redis) Setex(key, value string, seconds int) error {
	conn, err := getRedis(s)
	if err != nil {
		return err
	}

	return conn.Set(key, value, time.Duration(seconds)*time.Second).Err()
}

func (s *Redis) Setnx(key, value string) (bool, error) {
	conn, err := getRedis(s)
	if err != nil {
		return false, err
	}

	return conn.SetNX(key, value, 0).Result()
}

func (s *Redis) SetnxEx(key, value string, seconds int) (bool, error) {
	conn, err := getRedis(s)
	if err != nil {
		return false, err
	}

	return conn.SetNX(key, value, time.Duration(seconds)*time.Second).Result()
}

func (s *Redis) Sismember(key string, value interface{}) (bool, error) {
	conn, err := getRedis(s)
	if err != nil {
		return false, err
	}

	return conn.SIsMember(key, value).Result()
}

func (s *Redis) Srem(key string, values ...interface{}) (int, error) {
	conn, err := getRedis(s)
	if err != nil {
		return 0, err
	}

	if val, err := conn.SRem(key, values...).Result(); err != nil {
		return 0, err
	} else {
		return int(val), nil
	}
}

func (s *Redis) Smembers(key string) ([]string, error) {
	conn, err := getRedis(s)
	if err != nil {
		return nil, err
	}

	return conn.SMembers(key).Result()
}

func (s *Redis) Spop(key string) (string, error) {
	conn, err := getRedis(s)
	if err != nil {
		return "", err
	}

	return conn.SPop(key).Result()
}

func (s *Redis) Srandmember(key string, count int) ([]string, error) {
	conn, err := getRedis(s)
	if err != nil {
		return nil, err
	}

	return conn.SRandMemberN(key, int64(count)).Result()
}

func (s *Redis) Ttl(key string) (int, error) {
	conn, err := getRedis(s)
	if err != nil {
		return 0, err
	}

	if duration, err := conn.TTL(key).Result(); err != nil {
		return 0, err
	} else {
		return int(duration / time.Second), nil
	}
}

func (s *Redis) Zadd(key string, score int64, value string) (bool, error) {
	conn, err := getRedis(s)
	if err != nil {
		return false, err
	}

	if val, err := conn.ZAdd(key, red.Z{
		Score:  float64(score),
		Member: value,
	}).Result(); err != nil {
		return false, err
	} else {
		return val == 1, nil
	}
}

func (s *Redis) Zcard(key string) (int, error) {
	conn, err := getRedis(s)
	if err != nil {
		return 0, err
	}

	if val, err := conn.ZCard(key).Result(); err != nil {
		return 0, err
	} else {
		return int(val), nil
	}
}

func (s *Redis) Zcount(key string, start, stop int64) (int, error) {
	conn, err := getRedis(s)
	if err != nil {
		return 0, err
	}

	if val, err := conn.ZCount(key, strconv.FormatInt(start, 10),
		strconv.FormatInt(stop, 10)).Result(); err != nil {
		return 0, err
	} else {
		return int(val), nil
	}
}

func (s *Redis) Zincrby(key string, increment int64, field string) (int64, error) {
	conn, err := getRedis(s)
	if err != nil {
		return 0, err
	}

	if val, err := conn.ZIncrBy(key, float64(increment), field).Result(); err != nil {
		return 0, err
	} else {
		return int64(val), nil
	}
}

func (s *Redis) Zscore(key string, value string) (int64, error) {
	conn, err := getRedis(s)
	if err != nil {
		return 0, err
	}

	if val, err := conn.ZScore(key, value).Result(); err != nil {
		return 0, err
	} else {
		return int64(val), nil
	}
}

func (s *Redis) Zrank(key, field string) (int64, error) {
	conn, err := getRedis(s)
	if err != nil {
		return 0, err
	}

	return conn.ZRank(key, field).Result()
}

func (s *Redis) Zrem(key string, values ...interface{}) (int, error) {
	conn, err := getRedis(s)
	if err != nil {
		return 0, err
	}

	if val, err := conn.ZRem(key, values...).Result(); err != nil {
		return 0, err
	} else {
		return int(val), nil
	}
}

func (s *Redis) Zremrangebyscore(key string, start, stop int64) (int, error) {
	conn, err := getRedis(s)
	if err != nil {
		return 0, err
	}

	if val, err := conn.ZRemRangeByScore(key, strconv.FormatInt(start, 10),
		strconv.FormatInt(stop, 10)).Result(); err != nil {
		return 0, err
	} else {
		return int(val), nil
	}
}

func (s *Redis) Zremrangebyrank(key string, start, stop int64) (int, error) {
	conn, err := getRedis(s)
	if err != nil {
		return 0, err
	}

	if val, err := conn.ZRemRangeByRank(key, start, stop).Result(); err != nil {
		return 0, err
	} else {
		return int(val), nil
	}
}

func (s *Redis) Zrange(key string, start, stop int64) ([]string, error) {
	conn, err := getRedis(s)
	if err != nil {
		return nil, err
	}

	return conn.ZRange(key, start, stop).Result()
}

func (s *Redis) ZrangeWithScores(key string, start, stop int64) ([]Pair, error) {
	conn, err := getRedis(s)
	if err != nil {
		return nil, err
	}

	if vals, err := conn.ZRangeWithScores(key, start, stop).Result(); err != nil {
		return nil, err
	} else {
		return toPairs(vals), nil
	}
}

func (s *Redis) ZrangebyscoreWithScores(key string, start, stop int64) ([]Pair, error) {
	conn, err := getRedis(s)
	if err != nil {
		return nil, err
	}

	if vals, err := conn.ZRangeByScoreWithScores(key, red.ZRangeBy{
		Min: strconv.FormatInt(start, 10),
		Max: strconv.FormatInt(stop, 10),
	}).Result(); err != nil {
		return nil, err
	} else {
		return toPairs(vals), nil
	}
}

func (s *Redis) ZrangebyscoreWithScoresAndLimit(key string, start, stop int64, page, size int) ([]Pair, error) {
	if size <= 0 {
		return nil, nil
	}

	conn, err := getRedis(s)
	if err != nil {
		return nil, err
	}

	if vals, err := conn.ZRangeByScoreWithScores(key, red.ZRangeBy{
		Min:    strconv.FormatInt(start, 10),
		Max:    strconv.FormatInt(stop, 10),
		Offset: int64(page * size),
		Count:  int64(size),
	}).Result(); err != nil {
		return nil, err
	} else {
		return toPairs(vals), nil
	}
}

func (s *Redis) Zrevrange(key string, start, stop int64) ([]string, error) {
	conn, err := getRedis(s)
	if err != nil {
		return nil, err
	}

	return conn.ZRevRange(key, start, stop).Result()
}

func (s *Redis) ZrevrangebyscoreWithScores(key string, start, stop int64) ([]Pair, error) {
	conn, err := getRedis(s)
	if err != nil {
		return nil, err
	}

	if vals, err := conn.ZRevRangeByScoreWithScores(key, red.ZRangeBy{
		Min: strconv.FormatInt(start, 10),
		Max: strconv.FormatInt(stop, 10),
	}).Result(); err != nil {
		return nil, err
	} else {
		return toPairs(vals), nil
	}
}

func (s *Redis) ZrevrangebyscoreWithScoresAndLimit(key string, start, stop int64, page, size int) ([]Pair, error) {
	if size <= 0 {
		return nil, nil
	}

	conn, err := getRedis(s)
	if err != nil {
		return nil, err
	}

	if vals, err := conn.ZRevRangeByScoreWithScores(key, red.ZRangeBy{
		Min:    strconv.FormatInt(start, 10),
		Max:    strconv.FormatInt(stop, 10),
		Offset: int64(page * size),
		Count:  int64(size),
	}).Result(); err != nil {
		return nil, err
	} else {
		return toPairs(vals), nil
	}
}

func (s *Redis) scriptLoad(script string) (string, error) {
	conn, err := getRedis(s)
	if err != nil {
		return "", err
	}

	return conn.ScriptLoad(script).Result()
}

func getRedis(r *Redis) (RedisNode, error) {
	switch r.RedisType {
	case ClusterType:
		return GetRedisCluster(r.RedisAddr, r.RedisPass)
	case NodeType:
		return GetRedisClient(r.RedisAddr, r.RedisPass)
	default:
		return nil, fmt.Errorf("redis type '%s' is not supported", r.RedisType)
	}
}

func toPairs(vals []red.Z) []Pair {
	pairs := make([]Pair, len(vals))
	for i, val := range vals {
		switch member := val.Member.(type) {
		case string:
			pairs[i] = Pair{
				Key:   member,
				Score: int64(val.Score),
			}
		default:
			pairs[i] = Pair{
				Key:   fmt.Sprint(val.Member),
				Score: int64(val.Score),
			}
		}
	}
	return pairs
}

func toStrings(vals []interface{}) []string {
	ret := make([]string, len(vals))
	for i, val := range vals {
		if val == nil {
			ret[i] = ""
		} else {
			switch val := val.(type) {
			case string:
				ret[i] = val
			default:
				ret[i] = fmt.Sprint(val)
			}
		}
	}
	return ret
}
