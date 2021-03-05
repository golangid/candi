package database

import (
	"context"
	"database/sql"
	"log"
	"net/url"

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

	if env.BaseEnv().DbSQLReadDSN != "" {
		u, err := url.Parse(env.BaseEnv().DbSQLReadDSN)
		if err != nil {
			log.Panicf("SQL Read URL: %v", err)
		}
		inst.read, err = sql.Open(u.Scheme, env.BaseEnv().DbSQLReadDSN)
		if err != nil {
			log.Panicf("SQL Read: %v", err)
		}

		if err = inst.read.Ping(); err != nil {
			log.Panicf("SQL Read: %v", err)
		}
	}

	if env.BaseEnv().DbSQLWriteDSN != "" {
		u, err := url.Parse(env.BaseEnv().DbSQLWriteDSN)
		if err != nil {
			log.Panicf("SQL Write URL: %v", err)
		}
		inst.write, err = sql.Open(u.Scheme, env.BaseEnv().DbSQLWriteDSN)
		if err != nil {
			log.Panicf("SQL Write: %v", err)
		}

		if err = inst.write.Ping(); err != nil {
			log.Panicf("SQL Write: %v", err)
		}
	}

	return inst
}
