package interfaces

import (
	"context"
	"database/sql"

	"github.com/gomodule/redigo/redis"
	"go.mongodb.org/mongo-driver/mongo"
)

// SQLDatabase abstraction
type SQLDatabase interface {
	ReadDB() *sql.DB
	WriteDB() *sql.DB
	Disconnect()
}

// MongoDatabase abstraction
type MongoDatabase interface {
	ReadDB() *mongo.Database
	WriteDB() *mongo.Database
	Disconnect(ctx context.Context)
}

// RedisPool abstraction
type RedisPool interface {
	ReadPool() *redis.Pool
	WritePool() *redis.Pool
	Disconnect()
}
