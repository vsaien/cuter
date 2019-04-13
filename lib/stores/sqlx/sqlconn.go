package sqlx

import "database/sql"

type (
	// Session stands for raw connections or transaction sessions
	Session interface {
		Exec(query string, args ...interface{}) (sql.Result, error)
		Prepare(query string) (StmtSession, error)
		QueryRow(v interface{}, query string, args ...interface{}) error
		QueryRows(v interface{}, query string, args ...interface{}) error
	}

	// SqlConn only stands for raw connections, so Transact method can be called.
	SqlConn interface {
		Session
		Transact(func(session Session) error) error
	}

	StmtSession interface {
		Close() error
		Exec(args ...interface{}) (sql.Result, error)
		QueryRow(v interface{}, args ...interface{}) error
		QueryRows(v interface{}, args ...interface{}) error
	}

	// thread-safe
	// Because CORBA doesn't support PREPARE, so we need to combine the
	// query arguments into one string and do underlying query without arguments
	commonSqlConn struct {
		driverName string
		datasource string
		beginTx    beginnable
	}

	sessionConn interface {
		Exec(query string, args ...interface{}) (sql.Result, error)
		Query(query string, args ...interface{}) (*sql.Rows, error)
	}

	statement struct {
		stmt *sql.Stmt
	}

	stmtConn interface {
		Exec(args ...interface{}) (sql.Result, error)
		Query(args ...interface{}) (*sql.Rows, error)
	}
)

func (db *commonSqlConn) Exec(q string, args ...interface{}) (sql.Result, error) {
	conn, err := getSqlConn(db.driverName, db.datasource)
	if err != nil {
		logInstanceError(db.datasource, err)
		return nil, err
	}

	return exec(conn, q, args...)
}

func (db *commonSqlConn) Prepare(query string) (StmtSession, error) {
	conn, err := getSqlConn(db.driverName, db.datasource)
	if err != nil {
		logInstanceError(db.datasource, err)
		return nil, err
	}

	if stmt, err := conn.Prepare(query); err != nil {
		return nil, err
	} else {
		return statement{
			stmt: stmt,
		}, nil
	}
}

func (db *commonSqlConn) QueryRow(v interface{}, q string, args ...interface{}) error {
	return db.queryRows(func(rows *sql.Rows) error {
		return UnmarshalRow(v, rows)
	}, q, args...)
}

func (db *commonSqlConn) QueryRows(v interface{}, q string, args ...interface{}) error {
	return db.queryRows(func(rows *sql.Rows) error {
		return UnmarshalRows(v, rows)
	}, q, args...)
}

func (db *commonSqlConn) Transact(fn func(Session) error) (err error) {
	return transact(db, db.beginTx, fn)
}

func (db *commonSqlConn) queryRows(scanner func(*sql.Rows) error, q string, args ...interface{}) error {
	conn, err := getSqlConn(db.driverName, db.datasource)
	if err != nil {
		logInstanceError(db.datasource, err)
		return err
	}

	return query(conn, scanner, q, args...)
}

func (s statement) Close() error {
	return s.stmt.Close()
}

func (s statement) Exec(args ...interface{}) (sql.Result, error) {
	return execStmt(s.stmt, args...)
}

func (s statement) QueryRow(v interface{}, args ...interface{}) error {
	return queryStmt(s.stmt, func(rows *sql.Rows) error {
		return UnmarshalRow(v, rows)
	}, args...)
}

func (s statement) QueryRows(v interface{}, args ...interface{}) error {
	return queryStmt(s.stmt, func(rows *sql.Rows) error {
		return UnmarshalRows(v, rows)
	}, args...)
}
