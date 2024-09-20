package database

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"strings"

	"github.com/golangid/candi/config/env"
	"github.com/golangid/candi/logger"
)

type SQLInstance struct {
	DBRead, DBWrite *sql.DB
}

func (s *SQLInstance) ReadDB() *sql.DB {
	return s.DBRead
}

func (s *SQLInstance) WriteDB() *sql.DB {
	return s.DBWrite
}

func (s *SQLInstance) Health() map[string]error {
	mErr := make(map[string]error)
	if s.DBRead != nil {
		mErr["sql_read"] = s.DBRead.Ping()
	}
	if s.DBWrite != nil {
		mErr["sql_write"] = s.DBWrite.Ping()
	}
	return mErr
}

func (s *SQLInstance) Disconnect(ctx context.Context) (err error) {
	defer logger.LogWithDefer("\x1b[33;5msql\x1b[0m: disconnect...")()

	if s.DBRead != nil {
		if err := s.DBRead.Close(); err != nil {
			return err
		}
	}
	if s.DBWrite != nil {
		err = s.DBWrite.Close()
	}
	return
}

func (s *SQLInstance) Close() (err error) {
	return s.Disconnect(context.Background())
}

type SQLDatabaseOption func(db *sql.DB)

// InitSQLDatabase return sql db read & write instance from environment:
// SQL_DB_READ_DSN, SQL_DB_WRITE_DSN
// if want to create single connection, use SQL_DB_WRITE_DSN and set empty for SQL_DB_READ_DSN
func InitSQLDatabase(opts ...SQLDatabaseOption) *SQLInstance {
	defer logger.LogWithDefer("Load SQL connection...")()

	connReadDSN, connWriteDSN := env.BaseEnv().DbSQLReadDSN, env.BaseEnv().DbSQLWriteDSN
	if connReadDSN == "" {
		db := ConnectSQLDatabase(connWriteDSN, opts...)
		return &SQLInstance{
			DBRead: db, DBWrite: db,
		}
	}

	return &SQLInstance{
		DBRead:  ConnectSQLDatabase(connReadDSN, opts...),
		DBWrite: ConnectSQLDatabase(connWriteDSN, opts...),
	}
}

// ParseSQLDSN parse sql dsn
func ParseSQLDSN(source string) (driverName string, dsn string) {
	sqlDriver, conn, ok := strings.Cut(source, "://")
	if !ok {
		panic("SQL DSN: invalid url format")
	}
	driverName = sqlDriver

	switch sqlDriver {
	case "mysql", "sqlite3":
		dsn = conn

	case "sqlserver":
		driverName = "mssql"
		fallthrough

	case "postgres":
		if i := strings.LastIndex(conn, "@"); i > 0 {
			if username, password, ok := strings.Cut(conn[:i], ":"); ok {
				conn = fmt.Sprintf("%s:%s@%s", url.QueryEscape(username), url.QueryEscape(password), conn[i+1:])
			}
		}
		dsn = fmt.Sprintf("%s://%s", sqlDriver, conn)
	}
	return
}

// ConnectSQLDatabase connect to sql database with dsn
func ConnectSQLDatabase(dsn string, opts ...SQLDatabaseOption) *sql.DB {
	db, err := sql.Open(ParseSQLDSN(dsn))
	if err != nil {
		panic(fmt.Sprintf("SQL Connection: %v", err))
	}
	if err = db.Ping(); err != nil {
		panic(fmt.Sprintf("SQL Ping: %v", err))
	}

	for _, opt := range opts {
		opt(db)
	}
	return db
}
