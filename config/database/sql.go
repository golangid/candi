package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
)

// InitSQLDatabase return sql db read & write instance
func InitSQLDatabase(ctx context.Context, isUse bool) (read, write *sql.DB) {
	if !isUse {
		return
	}

	var err error
	descriptor := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("SQL_DB_READ_HOST"), os.Getenv("SQL_DB_READ_USER"), os.Getenv("SQL_DB_READ_PASSWORD"), os.Getenv("SQL_DB_READ_NAME"))
	read, err = sql.Open(os.Getenv("SQL_DRIVER_NAME"), descriptor)
	if err != nil {
		panic("SQL Read: " + err.Error())
	}

	descriptor = fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("SQL_DB_WRITE_HOST"), os.Getenv("SQL_DB_WRITE_USER"), os.Getenv("SQL_DB_WRITE_PASSWORD"), os.Getenv("SQL_DB_WRITE_NAME"))
	write, err = sql.Open(os.Getenv("SQL_DRIVER_NAME"), descriptor)
	if err != nil {
		panic("SQL Write: " + err.Error())
	}

	log.Println("Success load SQL connection")
	return
}
