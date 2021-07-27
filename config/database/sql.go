package database

import (
	"context"
	"database/sql"
	"log"
	"net/url"
	"strings"

	"pkg.agungdp.dev/candi/codebase/interfaces"
	"pkg.agungdp.dev/candi/config/env"
	"pkg.agungdp.dev/candi/logger"
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
	deferFunc := logger.LogWithDefer("sql: disconnect...")
	defer deferFunc()

	if err := s.read.Close(); err != nil {
		return err
	}
	return s.write.Close()
}

// InitSQLDatabase return sql db read & write instance from environment:
// SQL_DB_READ_DSN, SQL_DB_WRITE_DSN
func InitSQLDatabase() interfaces.SQLDatabase {
	deferFunc := logger.LogWithDefer("Load SQL connection...")
	defer deferFunc()

	return &sqlInstance{
		read:  ConnectSQLDatabase(env.BaseEnv().DbSQLReadDSN),
		write: ConnectSQLDatabase(env.BaseEnv().DbSQLWriteDSN),
	}
}

// ConnectSQLDatabase connect to sql database with dsn
func ConnectSQLDatabase(dsn string) *sql.DB {
	var sqlDriver string
	delimiter := "://"
	if u, err := url.Parse(dsn); err != nil {
		idx := strings.Index(dsn, delimiter)
		sqlDriver = dsn[0:idx]
		dsn = dsn[idx+len(delimiter):]
	} else {
		sqlDriver = u.Scheme
	}
	db, err := sql.Open(sqlDriver, dsn)
	if err != nil {
		log.Panicf("SQL Connection: %v", err)
	}
	if err = db.Ping(); err != nil {
		log.Panicf("SQL Connection: %v", err)
	}

	return db
}
