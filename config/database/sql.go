package database

import (
	"context"
	"database/sql"

	"pkg.agungdwiprasetyo.com/candi/codebase/interfaces"
	"pkg.agungdwiprasetyo.com/candi/config/env"
	"pkg.agungdwiprasetyo.com/candi/logger"
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
// SQL_DRIVER_NAME, SQL_DB_READ_HOST, SQL_DB_READ_USER, SQL_DB_READ_PASSWORD, SQL_DB_WRITE_HOST, SQL_DB_WRITE_USER, SQL_DB_WRITE_PASSWORD, SQL_DATABASE_NAME
func InitSQLDatabase() interfaces.SQLDatabase {
	deferFunc := logger.LogWithDefer("Load SQL connection...")
	defer deferFunc()

	inst := new(sqlInstance)

	var err error
	inst.read, err = sql.Open(env.BaseEnv().DbSQLDriver, env.BaseEnv().DbSQLReadDSN)
	if err != nil {
		panic("SQL Read: " + err.Error())
	}

	inst.write, err = sql.Open(env.BaseEnv().DbSQLDriver, env.BaseEnv().DbSQLWriteDSN)
	if err != nil {
		panic("SQL Write: " + err.Error())
	}

	return inst
}
