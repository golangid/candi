package database

import (
	"context"
	"fmt"
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
	deferFunc := logger.LogWithDefer("mongodb: disconnect...")
	defer deferFunc()

	if err := m.write.Client().Disconnect(ctx); err != nil {
		return err
	}
	return m.read.Client().Disconnect(ctx)
}

// InitMongoDB return mongo db read & write instance from environment:
// MONGODB_HOST_WRITE, MONGODB_HOST_READ
func InitMongoDB(ctx context.Context) interfaces.MongoDatabase {
	deferFunc := logger.LogWithDefer("Load MongoDB connection...")
	defer deferFunc()

	mi := &mongoInstance{}
	if env.BaseEnv().DbMongoReadHost != "" {
		mi.read = ConnectMongoDB(ctx, env.BaseEnv().DbMongoReadHost)
	}
	if env.BaseEnv().DbMongoWriteHost != "" {
		mi.write = ConnectMongoDB(ctx, env.BaseEnv().DbMongoWriteHost)
	}
	return mi
}

// ConnectMongoDB connect to mongodb with dsn
func ConnectMongoDB(ctx context.Context, dsn string) *mongo.Database {
	clientOpts := []*options.ClientOptions{
		options.Client().SetConnectTimeout(10 * time.Second),
		options.Client().SetServerSelectionTimeout(10 * time.Second),
	}

	// get mongo dsn from env
	connDSN, err := connstring.ParseAndValidate(dsn)
	if err != nil {
		panic(err)
	}
	// connect to MongoDB
	client, err := mongo.NewClient(append([]*options.ClientOptions{options.Client().ApplyURI(connDSN.String())}, clientOpts...)...)
	if err != nil {
		panic(fmt.Sprintf("mongodb: %v, conn: %s", err, connDSN.String()))
	}
	if err := client.Connect(ctx); err != nil {
		panic(fmt.Sprintf("mongodb error connect: %v", err))
	}
	if err = client.Ping(ctx, readpref.Primary()); err != nil {
		panic(fmt.Sprintf("mongodb ping: %v", err))
	}

	return client.Database(connDSN.Database)
}
