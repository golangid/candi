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

	inst := new(sqlInstance)
	var sqlDriver string
	var err error
	delimiter := "://"

	if env.BaseEnv().DbSQLReadDSN != "" {
		dsn := env.BaseEnv().DbSQLReadDSN
		if u, err := url.Parse(dsn); err != nil {
			idx := strings.Index(dsn, delimiter)
			sqlDriver = dsn[0:idx]
			dsn = dsn[idx+len(delimiter):]
		} else {
			sqlDriver = u.Scheme
		}
		inst.read, err = sql.Open(sqlDriver, dsn)
		if err != nil {
			log.Panicf("SQL Read: %v", err)
		}

		if err = inst.read.Ping(); err != nil {
			log.Panicf("SQL Read: %v", err)
		}
	}

	if env.BaseEnv().DbSQLWriteDSN != "" {
		dsn := env.BaseEnv().DbSQLReadDSN
		if u, err := url.Parse(dsn); err != nil {
			idx := strings.Index(dsn, delimiter)
			sqlDriver = env.BaseEnv().DbSQLReadDSN[0:idx]
			dsn = dsn[idx+len(delimiter):]
		} else {
			sqlDriver = u.Scheme
		}
		inst.write, err = sql.Open(sqlDriver, dsn)
		if err != nil {
			log.Panicf("SQL Write: %v", err)
		}

		if err = inst.write.Ping(); err != nil {
			log.Panicf("SQL Write: %v", err)
		}
	}

	return inst
}
