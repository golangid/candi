package database

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"pkg.agungdp.dev/candi/codebase/interfaces"
	"pkg.agungdp.dev/candi/config/env"
	"pkg.agungdp.dev/candi/logger"
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
// MONGODB_HOST_WRITE, MONGODB_HOST_READ, MONGODB_DATABASE_NAME
func InitMongoDB(ctx context.Context) interfaces.MongoDatabase {
	deferFunc := logger.LogWithDefer("Load MongoDB connection...")
	defer deferFunc()

	// create db instance
	dbInstance := new(mongoInstance)
	dbName := env.BaseEnv().DbMongoDatabaseName

	clientOpts := []*options.ClientOptions{
		options.Client().SetConnectTimeout(10 * time.Second),
		options.Client().SetServerSelectionTimeout(10 * time.Second),
	}

	// get write mongo from env
	hostWrite := env.BaseEnv().DbMongoWriteHost
	// connect to MongoDB
	client, err := mongo.NewClient(append([]*options.ClientOptions{options.Client().ApplyURI(hostWrite)}, clientOpts...)...)
	if err != nil {
		panic(fmt.Errorf("mongo: %v, conn: %s", err, hostWrite))
	}
	if err := client.Connect(ctx); err != nil {
		panic(fmt.Errorf("mongo write error connect: %v", err))
	}
	dbInstance.write = client.Database(dbName)

	// get read mongo from env
	hostRead := env.BaseEnv().DbMongoReadHost
	// connect to MongoDB
	client, err = mongo.NewClient(append([]*options.ClientOptions{options.Client().ApplyURI(hostRead)}, clientOpts...)...)
	if err != nil {
		panic(fmt.Errorf("mongo: %v, conn: %s", err, hostRead))
	}
	if err := client.Connect(ctx); err != nil {
		panic(fmt.Errorf("mongo read error connect: %v", err))
	}
	dbInstance.read = client.Database(dbName)

	return dbInstance
}
