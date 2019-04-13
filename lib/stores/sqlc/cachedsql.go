package sqlc

import (
	"database/sql"

	"github.com/vsaien/cuter/lib/stores/internal"
	"github.com/vsaien/cuter/lib/stores/redis"
	"github.com/vsaien/cuter/lib/stores/sqlx"
	"github.com/vsaien/cuter/lib/syncx"
)

var (
	ErrNotFound = internal.ErrNotFound

	exclusiveCalls = syncx.NewExclusiveCalls()
	stat           = internal.NewCacheStat("sqlc")
)

type (
	ExecFn  func(conn sqlx.Session) (sql.Result, error)
	QueryFn func(conn sqlx.Session, v interface{}) error

	CachedConn struct {
		db    sqlx.SqlConn
		cache internal.Cache
	}
)

func NewCachedConn(db sqlx.SqlConn, rds *redis.Redis) CachedConn {
	return CachedConn{
		db:    db,
		cache: internal.NewCache(rds, exclusiveCalls, &stat),
	}
}

func (cc CachedConn) DelCache(key string) error {
	return cc.cache.DelCache(key)
}

func (cc CachedConn) Exec(q string, args ...interface{}) (sql.Result, error) {
	return cc.db.Exec(q, args...)
}

func (cc CachedConn) ExecDropCache(exec ExecFn, key string) (sql.Result, error) {
	if err := cc.DelCache(key); err != nil {
		return nil, err
	}

	return exec(cc.db)
}

func (cc CachedConn) QueryRow(v interface{}, key string, seconds int, query QueryFn) error {
	return cc.cache.Take(v, key, seconds, func(v interface{}) error {
		if err := query(cc.db, v); err == sql.ErrNoRows {
			return internal.ErrNotFound
		} else {
			return err
		}
	})
}

func (cc CachedConn) QueryRows(v interface{}, q string, args ...interface{}) error {
	return cc.db.QueryRows(v, q, args...)
}

func (cc CachedConn) SetCache(key string, v interface{}, seconds int) error {
	return cc.cache.SetCache(key, v, seconds)
}

func (cc CachedConn) Transact(fn func(sqlx.Session) error) error {
	return cc.db.Transact(fn)
}
