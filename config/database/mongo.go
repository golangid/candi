package database

import (
	"context"
	"fmt"
	"log"
	"os"

	"agungdwiprasetyo.com/backend-microservices/pkg/codebase/interfaces"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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
func (m *mongoInstance) Disconnect(ctx context.Context) error {
	defer log.Println("mongo: success disconnect")
	if err := m.write.Client().Disconnect(ctx); err != nil {
		return err
	}
	return m.read.Client().Disconnect(ctx)
}

// InitMongoDB return mongo db read & write instance
func InitMongoDB(ctx context.Context) interfaces.MongoDatabase {
	dbInstance := new(mongoInstance)

	dbName, ok := os.LookupEnv("MONGODB_DATABASE_NAME")
	if !ok {
		panic("missing MONGODB_DATABASE_NAME environment")
	}

	// init write mongodb
	hostWrite := os.Getenv("MONGODB_HOST_WRITE")
	client, err := mongo.NewClient(options.Client().ApplyURI(hostWrite))
	if err != nil {
		panic(fmt.Errorf("mongo: %v, conn: %s", err, hostWrite))
	}
	if err := client.Connect(ctx); err != nil {
		panic(fmt.Errorf("mongo: %v, conn: %s", err, hostWrite))
	}
	dbInstance.write = client.Database(dbName)

	// init read mongodb
	hostRead := os.Getenv("MONGODB_HOST_READ")
	client, err = mongo.NewClient(options.Client().ApplyURI(hostRead))
	if err != nil {
		panic(fmt.Errorf("mongo: %v, conn: %s", err, hostRead))
	}
	if err := client.Connect(ctx); err != nil {
		panic(fmt.Errorf("mongo: %v, conn: %s", err, hostRead))
	}
	dbInstance.read = client.Database(dbName)

	log.Println("Success load Mongo connection")
	return dbInstance
}
