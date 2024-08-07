package database

import (
	"context"
	"log"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.mongodb.org/mongo-driver/x/mongo/driver/connstring"

	"github.com/golangid/candi/codebase/interfaces"
	"github.com/golangid/candi/config/env"
	"github.com/golangid/candi/logger"
)

type mongoInstance struct {
	read, write *mongo.Database
}

func (m *mongoInstance) ReadDB() *mongo.Database {
	return m.read
}
func (m *mongoInstance) WriteDB() *mongo.Database {
	return m.write
}
func (m *mongoInstance) Health() map[string]error {
	var readErr, writeErr error
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		readErr = m.read.Client().Ping(context.Background(), readpref.Primary())
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		writeErr = m.write.Client().Ping(context.Background(), readpref.Primary())
	}()
	wg.Wait()
	return map[string]error{
		"mongo_read": readErr, "mongo_write": writeErr,
	}
}
func (m *mongoInstance) Disconnect(ctx context.Context) (err error) {
	defer logger.LogWithDefer("\x1b[33;5mmongodb\x1b[0m: disconnect...")()

	if err := m.write.Client().Disconnect(ctx); err != nil {
		return err
	}
	return m.read.Client().Disconnect(ctx)
}

// InitMongoDB return mongo db read & write instance from environment:
// MONGODB_HOST_WRITE, MONGODB_HOST_READ
// if want to create single connection, use MONGODB_HOST_WRITE and set empty for MONGODB_HOST_READ
func InitMongoDB(ctx context.Context, opts ...*options.ClientOptions) interfaces.MongoDatabase {
	defer logger.LogWithDefer("Load MongoDB connection...")()

	connReadDSN, connWriteDSN := env.BaseEnv().DbMongoReadHost, env.BaseEnv().DbMongoWriteHost
	if connReadDSN == "" {
		db := ConnectMongoDB(ctx, connWriteDSN, opts...)
		return &mongoInstance{read: db, write: db}
	}

	return &mongoInstance{
		read:  ConnectMongoDB(ctx, connReadDSN, opts...),
		write: ConnectMongoDB(ctx, connWriteDSN, opts...),
	}
}

// ConnectMongoDB connect to mongodb with dsn
func ConnectMongoDB(ctx context.Context, dsn string, opts ...*options.ClientOptions) *mongo.Database {
	connDSN, err := connstring.ParseAndValidate(dsn)
	if err != nil {
		log.Panic(err)
	}

	clientOpts := []*options.ClientOptions{
		options.Client().ApplyURI(connDSN.String()),
		options.Client().SetConnectTimeout(10 * time.Second),
		options.Client().SetServerSelectionTimeout(10 * time.Second),
	}
	clientOpts = append(clientOpts, opts...)

	client, err := mongo.Connect(ctx, clientOpts...)
	if err != nil {
		log.Panicf("mongodb: %v, conn: %s", err, connDSN.String())
	}
	if err = client.Ping(ctx, readpref.Primary()); err != nil {
		log.Panicf("mongodb ping: %v", err)
	}

	return client.Database(connDSN.Database)
}
