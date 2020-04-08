package config

import (
	"context"
	"crypto/rsa"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/agungdwiprasetyo/backend-microservices/config/database"
	"github.com/agungdwiprasetyo/backend-microservices/config/key"
	"github.com/gomodule/redigo/redis"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
)

// Config app
type Config struct {
	MongoRead, MongoWrite         *mongo.Database
	SQLRead, SQLWrite             *sql.DB
	RedisReadPool, RedisWritePool *redis.Pool
	PrivateKey                    *rsa.PrivateKey
	PublicKey                     *rsa.PublicKey
}

// Env model
type Env struct {
	RootApp string

	// UseGraphQL env
	UseGraphQL bool
	// UseGRPC env
	UseGRPC bool

	// Development env checking, this env for debug purpose
	Development string

	// HTTPPort config
	HTTPPort uint16
	// GRPCPort Config
	GRPCPort uint16

	// BasicAuthUsername config
	BasicAuthUsername string
	// BasicAuthPassword config
	BasicAuthPassword string

	// GRPC auth key
	GRPCAuthKey string

	// CacheExpired config
	CacheExpired time.Duration
}

// GlobalEnv global environment
var GlobalEnv Env

// Init app config
func Init(ctx context.Context, rootApp string) *Config {
	loadEnv(rootApp)

	cfgChan := make(chan *Config)
	errConnect := make(chan interface{})
	go func() {
		defer func() {
			if r := recover(); r != nil {
				errConnect <- r
			}
			close(cfgChan)
			close(errConnect)
		}()

		var cfg Config
		cfg.MongoRead, cfg.MongoWrite = database.InitMongoDB(ctx)
		cfg.RedisReadPool, cfg.RedisWritePool = database.InitRedis()
		cfg.PrivateKey = key.LoadPrivateKey()
		cfg.PublicKey = key.LoadPublicKey()

		cfgChan <- &cfg
	}()

	// with timeout to init configuration
	select {
	case cfg := <-cfgChan:
		return cfg
	case <-ctx.Done():
		panic(fmt.Errorf("Timeout to init configuration: %v", ctx.Err()))
	case e := <-errConnect:
		panic(fmt.Errorf("Failed init configuration :=> %v", e))
	}
}

func loadEnv(rootApp string) {
	// load .env
	err := godotenv.Load(rootApp + "/.env")
	if err != nil {
		log.Println(err)
		panic(errors.New(".env is not loaded properly"))
	}

	os.Setenv("APP_PATH", rootApp)
	GlobalEnv.RootApp = rootApp

	// ------------------------------------
	useGraphQL, ok := os.LookupEnv("USE_GRAPHQL")
	if !ok {
		panic("missing USE_GRAPHQL environment")
	}
	GlobalEnv.UseGraphQL, _ = strconv.ParseBool(useGraphQL)

	useGRPC, ok := os.LookupEnv("USE_GRPC")
	if !ok {
		panic("missing USE_GRPC environment")
	}
	GlobalEnv.UseGRPC, _ = strconv.ParseBool(useGRPC)

	// ------------------------------------

	if port, err := strconv.Atoi(os.Getenv("HTTP_PORT")); err != nil {
		panic("missing HTTP_PORT environment")
	} else {
		GlobalEnv.HTTPPort = uint16(port)
	}

	if port, err := strconv.Atoi(os.Getenv("GRPC_PORT")); err != nil {
		panic("missing GRPC_PORT environment")
	} else {
		GlobalEnv.GRPCPort = uint16(port)
	}

	// ------------------------------------
	GlobalEnv.BasicAuthUsername, ok = os.LookupEnv("BASIC_AUTH_USERNAME")
	if !ok {
		panic("missing BASIC_AUTH_USERNAME environment")
	}
	GlobalEnv.BasicAuthPassword, ok = os.LookupEnv("BASIC_AUTH_PASS")
	if !ok {
		panic("missing BASIC_AUTH_PASS environment")
	}

	GlobalEnv.GRPCAuthKey, ok = os.LookupEnv("GRPC_AUTH_KEY")
	if !ok {
		panic("missing GRPC_AUTH_KEY environment")
	}
}

// Exit release all connection, think as deferred function in main
func (c *Config) Exit(ctx context.Context) {

	log.Println("\x1b[33;1mConfig: Success close all connection\x1b[0m")
}
