package database

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	"pkg.agungdwiprasetyo.com/gendon/codebase/interfaces"
	"pkg.agungdwiprasetyo.com/gendon/logger"
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
func (s *sqlInstance) Disconnect(ctx context.Context) (err error) {
	deferFunc := logger.LogWithDefer("sql: disconnect......")
	defer deferFunc()

	if err := s.read.Close(); err != nil {
		return err
	}
	return s.write.Close()
}

// InitSQLDatabase return sql db read & write instance
func InitSQLDatabase() interfaces.SQLDatabase {
	deferFunc := logger.LogWithDefer("Load SQL connection...")
	defer deferFunc()

	inst := new(sqlInstance)

	dbName, ok := os.LookupEnv("SQL_DATABASE_NAME")
	if !ok {
		panic("missing SQL_DATABASE_NAME environment")
	}

	var err error
	descriptor := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("SQL_DB_READ_HOST"), os.Getenv("SQL_DB_READ_USER"), os.Getenv("SQL_DB_READ_PASSWORD"), dbName)
	inst.read, err = sql.Open(os.Getenv("SQL_DRIVER_NAME"), descriptor)
	if err != nil {
		panic("SQL Read: " + err.Error())
	}

	descriptor = fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("SQL_DB_WRITE_HOST"), os.Getenv("SQL_DB_WRITE_USER"), os.Getenv("SQL_DB_WRITE_PASSWORD"), dbName)
	inst.write, err = sql.Open(os.Getenv("SQL_DRIVER_NAME"), descriptor)
	if err != nil {
		panic("SQL Write: " + err.Error())
	}

	return inst
}
