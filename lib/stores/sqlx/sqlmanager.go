package sqlx

import (
	"database/sql"
	"sync"
	"time"

	"github.com/vsaien/cuter/lib/syncx"
)

const (
	maxIdleConns = 64
	maxOpenConns = 64
	maxLifetime  = time.Minute
)

var sqlManager *SqlManager

type (
	pingedDB struct {
		db   *sql.DB
		once sync.Once
	}

	SqlManager struct {
		conns          map[string]*pingedDB
		exclusiveCalls syncx.ExclusiveCalls
	}
)

func init() {
	sqlManager = &SqlManager{
		conns:          make(map[string]*pingedDB),
		exclusiveCalls: syncx.NewExclusiveCalls(),
	}
}

func getCachedSqlConn(driverName, server string) (*pingedDB, error) {
	val, err := sqlManager.exclusiveCalls.Do(server, func() (interface{}, error) {
		pdb, ok := sqlManager.conns[server]
		if ok {
			return pdb, nil
		}

		conn, err := newDBConnection(driverName, server)
		if err != nil {
			return nil, err
		}

		pdb = &pingedDB{
			db: conn,
		}
		sqlManager.conns[server] = pdb

		return pdb, nil
	})
	if err != nil {
		return nil, err
	}

	return val.(*pingedDB), nil
}

func getSqlConn(driverName, server string) (*sql.DB, error) {
	pdb, err := getCachedSqlConn(driverName, server)
	if err != nil {
		return nil, err
	}

	pdb.once.Do(func() {
		err = pdb.db.Ping()
	})
	if err != nil {
		return nil, err
	}

	return pdb.db, nil
}

func newDBConnection(driverName, datasource string) (*sql.DB, error) {
	conn, err := sql.Open(driverName, datasource)
	if err != nil {
		return nil, err
	}

	// we need to do this until the issue https://github.com/golang/go/issues/9851 get fixed
	// discussed here https://github.com/go-sql-driver/mysql/issues/257
	// if the discussed SetMaxIdleTimeout methods added, we'll change this behavior
	// 8 means we can't have more than 8 goroutines to concurrently access the same database.
	conn.SetMaxIdleConns(maxIdleConns)
	conn.SetMaxOpenConns(maxOpenConns)
	conn.SetConnMaxLifetime(maxLifetime)

	return conn, nil
}
