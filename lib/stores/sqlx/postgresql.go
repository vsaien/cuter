package sqlx

import _ "github.com/lib/pq"

const postgreDriverName = "postgres"

func NewPostgre(datasource string) SqlConn {
	return &commonSqlConn{
		driverName: postgreDriverName,
		datasource: datasource,
		beginTx:    beginStd,
	}
}
