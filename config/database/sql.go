package database

import (
	"context"
	"database/sql"
	"fmt"
	"os"
)

// InitSQLDatabase return sql db read & write instance
func InitSQLDatabase(ctx context.Context, driverName string) (read, write *sql.DB) {
	var err error
	descriptor := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("SQL_DB_READ_HOST"), os.Getenv("SQL_DB_READ_USER"), os.Getenv("SQL_DB_READ_PASSWORD"), os.Getenv("SQL_DB_READ_NAME"))
	read, err = sql.Open(driverName, descriptor)
	if err != nil {
		panic(err)
	}

	descriptor = fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("SQL_DB_WRITE_HOST"), os.Getenv("SQL_DB_WRITE_USER"), os.Getenv("SQL_DB_WRITE_PASSWORD"), os.Getenv("SQL_DB_WRITE_NAME"))
	write, err = sql.Open(driverName, descriptor)
	if err != nil {
		panic(err)
	}

	return
}
