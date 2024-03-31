package database

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"strings"

	"github.com/golangid/candi/codebase/interfaces"
	"github.com/golangid/candi/config/env"
	"github.com/golangid/candi/logger"
)

type sqlInstance struct {
	read, write *sql.DB
}

func (s *sqlInstance) ReadDB() *sql.DB {
	return s.read
}
func (s *sqlInstance) WriteDB() *sql.DB {
	return s.write
}
func (s *sqlInstance) Health() map[string]error {
	mErr := make(map[string]error)
	mErr["sql_read"] = s.read.Ping()
	mErr["sql_write"] = s.write.Ping()
	return mErr
}
func (s *sqlInstance) Disconnect(ctx context.Context) (err error) {
	defer logger.LogWithDefer("sql: disconnect...")()

	if err := s.read.Close(); err != nil {
		return err
	}
	return s.write.Close()
}

type SQLDatabaseOption func(db *sql.DB)

// InitSQLDatabase return sql db read & write instance from environment:
// SQL_DB_READ_DSN, SQL_DB_WRITE_DSN
// if want to create single connection, use SQL_DB_WRITE_DSN and set empty for SQL_DB_READ_DSN
func InitSQLDatabase(opts ...SQLDatabaseOption) interfaces.SQLDatabase {
	defer logger.LogWithDefer("Load SQL connection...")()

	connReadDSN, connWriteDSN := env.BaseEnv().DbSQLReadDSN, env.BaseEnv().DbSQLWriteDSN
	if connReadDSN == "" {
		db := ConnectSQLDatabase(connWriteDSN, opts...)
		return &sqlInstance{
			read: db, write: db,
		}
	}

	return &sqlInstance{
		read:  ConnectSQLDatabase(connReadDSN, opts...),
		write: ConnectSQLDatabase(connWriteDSN, opts...),
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
