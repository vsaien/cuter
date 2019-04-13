package sqlx

import _ "github.com/go-sql-driver/mysql"

const (
	mysqlDriverName = "mysql"

	// because aliyun is using the corba mechanism, so we just use this name to communicate with underlying drivers.
	corbaSql = "corba"
)

type SqlOption func(*commonSqlConn)

func NewMysql(datasource string, opts ...SqlOption) SqlConn {
	conn := &commonSqlConn{
		driverName: mysqlDriverName,
		datasource: datasource,
		beginTx:    beginStd,
	}

	for _, opt := range opts {
		opt(conn)
	}

	return conn
}

func WithAliyun() SqlOption {
	return func(conn *commonSqlConn) {
		conn.beginTx = beginAliyun
	}
}
