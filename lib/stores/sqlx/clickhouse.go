package sqlx

import _ "github.com/kshvakov/clickhouse"

const clickHouseDriverName = "clickhouse"

func NewClickHouse(datasource string, opts ...SqlOption) SqlConn {
	conn := &commonSqlConn{
		driverName: clickHouseDriverName,
		datasource: datasource,
		beginTx:    beginStd,
	}

	for _, opt := range opts {
		opt(conn)
	}

	return conn
}
