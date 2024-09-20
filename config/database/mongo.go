package database

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.mongodb.org/mongo-driver/x/mongo/driver/connstring"

	"github.com/golangid/candi/config/env"
	"github.com/golangid/candi/logger"
)

type MongoInstance struct {
	DBRead, DBWrite *mongo.Database
}

func (m *MongoInstance) ReadDB() *mongo.Database {
	return m.DBRead
}

func (m *MongoInstance) WriteDB() *mongo.Database {
	return m.DBWrite
}

func (m *MongoInstance) Health() map[string]error {
	ctx := context.Background()
	mErr := make(map[string]error)
	if m.DBRead != nil {
		mErr["mongo_read"] = m.DBRead.Client().Ping(ctx, readpref.Primary())
	}
	if m.DBWrite != nil {
		mErr["mongo_write"] = m.DBWrite.Client().Ping(ctx, readpref.Primary())
	}
	return mErr
}

func (m *MongoInstance) Disconnect(ctx context.Context) (err error) {
	defer logger.LogWithDefer("\x1b[33;5mmongodb\x1b[0m: disconnect...")()

	if m.DBWrite != nil {
		if err := m.DBWrite.Client().Disconnect(ctx); err != nil {
			return err
		}
	}
	if m.DBRead != nil {
		err = m.DBRead.Client().Disconnect(ctx)
	}
	return
}

func (m *MongoInstance) Close() (err error) {
	return m.Disconnect(context.Background())
}

// InitMongoDB return mongo db read & write instance from environment:
// MONGODB_HOST_WRITE, MONGODB_HOST_READ
// if want to create single connection, use MONGODB_HOST_WRITE and set empty for MONGODB_HOST_READ
func InitMongoDB(ctx context.Context, opts ...*options.ClientOptions) *MongoInstance {
	defer logger.LogWithDefer("Load MongoDB connection...")()

	connReadDSN, connWriteDSN := env.BaseEnv().DbMongoReadHost, env.BaseEnv().DbMongoWriteHost
	if connReadDSN == "" {
		db := ConnectMongoDB(ctx, connWriteDSN, opts...)
		return &MongoInstance{DBRead: db, DBWrite: db}
	}

	return &MongoInstance{
		DBRead:  ConnectMongoDB(ctx, connReadDSN, opts...),
		DBWrite: ConnectMongoDB(ctx, connWriteDSN, opts...),
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
