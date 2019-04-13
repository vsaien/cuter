package sqlx

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/vsaien/cuter/lib/logx"
)

const slowThreshold = time.Millisecond * 500

func exec(conn sessionConn, q string, args ...interface{}) (sql.Result, error) {
	stmt, err := format(q, args...)
	if err != nil {
		return nil, err
	}

	startTime := time.Now()
	result, err := conn.Exec(q, args...)
	duration := time.Since(startTime)
	if duration > slowThreshold {
		logx.Slowf("[SQL] exec: slowcall(%s) - %s", duration, stmt)
	} else {
		logx.Infof("sql exec: %s - %s", duration, stmt)
	}
	if err != nil {
		logSqlError(stmt, err)
	}

	return result, err
}

func execStmt(conn stmtConn, args ...interface{}) (sql.Result, error) {
	stmt := fmt.Sprint(args...)
	startTime := time.Now()
	result, err := conn.Exec(args...)
	duration := time.Since(startTime)
	if duration > slowThreshold {
		logx.Slowf("[SQL] execStmt: slowcall(%s) - %s", duration, stmt)
	} else {
		logx.Infof("sql execStmt: %s - %s", duration, stmt)
	}
	if err != nil {
		logSqlError(stmt, err)
	}

	return result, err
}

func query(conn sessionConn, scanner func(*sql.Rows) error, q string, args ...interface{}) error {
	stmt, err := format(q, args...)
	if err != nil {
		return err
	}

	startTime := time.Now()
	rows, err := conn.Query(q, args...)
	duration := time.Since(startTime)
	if duration > slowThreshold {
		logx.Slowf("[SQL] query: slowcall(%s) - %s", duration, stmt)
	} else {
		logx.Infof("sql query: %s - %s", duration, stmt)
	}
	if err != nil {
		logSqlError(stmt, err)
		return err
	}
	defer rows.Close()

	return scanner(rows)
}

func queryStmt(conn stmtConn, scanner func(*sql.Rows) error, args ...interface{}) error {
	stmt := fmt.Sprint(args...)
	startTime := time.Now()
	rows, err := conn.Query(args...)
	duration := time.Since(startTime)
	if duration > slowThreshold {
		logx.Slowf("[SQL] queryStmt: slowcall(%s) - %s", duration, stmt)
	} else {
		logx.Infof("sql queryStmt: %s - %s", duration, stmt)
	}
	if err != nil {
		logSqlError(stmt, err)
		return err
	}
	defer rows.Close()

	return scanner(rows)
}
