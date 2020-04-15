package database

import (
	"context"
	"fmt"
	"log"
	"os"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// InitMongoDB return mongo db read & write instance
func InitMongoDB(ctx context.Context, isUse bool) (read *mongo.Database, write *mongo.Database) {
	if !isUse {
		return
	}

	// init write mongodb
	hostWrite := os.Getenv("MONGODB_HOST_WRITE")
	dbNameWrite := os.Getenv("MONGODB_NAME_WRITE")
	client, err := mongo.NewClient(options.Client().ApplyURI(hostWrite))
	if err != nil {
		panic(fmt.Errorf("mongo: %v, conn: %s", err, hostWrite))
	}
	if err := client.Connect(ctx); err != nil {
		panic(fmt.Errorf("mongo: %v, conn: %s", err, hostWrite))
	}
	write = client.Database(dbNameWrite)

	// init read mongodb
	hostRead := os.Getenv("MONGODB_HOST_READ")
	dbNameRead := os.Getenv("MONGODB_NAME_READ")
	client, err = mongo.NewClient(options.Client().ApplyURI(hostRead))
	if err != nil {
		panic(fmt.Errorf("mongo: %v, conn: %s", err, hostRead))
	}
	if err := client.Connect(ctx); err != nil {
		panic(fmt.Errorf("mongo: %v, conn: %s", err, hostRead))
	}
	read = client.Database(dbNameRead)

	log.Println("Success load Mongo connection")
	return
}
