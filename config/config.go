package config

import (
	"context"
	"crypto/rsa"
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"agungdwiprasetyo.com/backend-microservices/config/broker"
	"agungdwiprasetyo.com/backend-microservices/config/database"
	"agungdwiprasetyo.com/backend-microservices/config/key"
	"github.com/Shopify/sarama"
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
	KafkaConsumerConfig           *sarama.Config
}

// Env model
type Env struct {
	RootApp string

	// UseHTTP env
	UseHTTP bool
	// UseGraphQL env
	UseGraphQL bool
	// UseGRPC env
	UseGRPC bool
	// UseKafka env
	UseKafka bool

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

	Kafka struct {
		Brokers       []string
		ClientID      string
		ConsumerGroup string
	}
}

// GlobalEnv global environment
var GlobalEnv Env

// db connection
var useSQL, useMongo, useRedis, useRSAKey bool

// Init app config
func Init(ctx context.Context, rootApp string) *Config {
	loadBaseEnv(rootApp)

	cfgChan := make(chan *Config)
	errConnect := make(chan error)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				errConnect <- fmt.Errorf("Failed init configuration :=> %v", r)
			}
			close(cfgChan)
			close(errConnect)
		}()

		var cfg Config
		cfg.MongoRead, cfg.MongoWrite = database.InitMongoDB(ctx, useMongo)
		cfg.SQLRead, cfg.SQLWrite = database.InitSQLDatabase(ctx, useSQL)
		cfg.RedisReadPool, cfg.RedisWritePool = database.InitRedis(useRedis)
		cfg.PrivateKey, cfg.PublicKey = key.LoadRSAKey(useRSAKey)
		cfg.KafkaConsumerConfig = broker.InitKafkaConfig()

		cfgChan <- &cfg
	}()

	// with timeout to init configuration
	select {
	case cfg := <-cfgChan:
		return cfg
	case <-ctx.Done():
		panic(fmt.Errorf("Timeout to init configuration: %v", ctx.Err()))
	case e := <-errConnect:
		panic(e)
	}
}

func loadBaseEnv(appLocation string) {
	// load main .env and additional .env in app
	if err := godotenv.Overload(".env", appLocation+"/.env"); err != nil {
		log.Println(err)
	}

	rootApp, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	os.Setenv("APP_PATH", rootApp)
	GlobalEnv.RootApp = rootApp

	// ------------------------------------
	useHTTP, ok := os.LookupEnv("USE_HTTP")
	if !ok {
		panic("missing USE_HTTP environment")
	}
	GlobalEnv.UseHTTP, _ = strconv.ParseBool(useHTTP)

	useGraphQL, ok := os.LookupEnv("USE_GRAPHQL")
	if !ok {
		panic("missing USE_GRAPHQL environment")
	}
	GlobalEnv.UseGraphQL, _ = strconv.ParseBool(useGraphQL)
	if GlobalEnv.UseGraphQL && !GlobalEnv.UseHTTP {
		panic("GraphQL required http server")
	}

	useGRPC, ok := os.LookupEnv("USE_GRPC")
	if !ok {
		panic("missing USE_GRPC environment")
	}
	GlobalEnv.UseGRPC, _ = strconv.ParseBool(useGRPC)

	useKafka, ok := os.LookupEnv("USE_KAFKA")
	if !ok {
		panic("missing USE_KAFKA environment")
	}
	GlobalEnv.UseKafka, _ = strconv.ParseBool(useKafka)

	// ---------------------------
	useMongoEnv, ok := os.LookupEnv("USE_MONGO")
	if !ok {
		panic("missing USE_MONGO environment")
	}
	useMongo, _ = strconv.ParseBool(useMongoEnv)

	useSQLEnv, ok := os.LookupEnv("USE_SQL")
	if !ok {
		panic("missing USE_SQL environment")
	}
	useSQL, _ = strconv.ParseBool(useSQLEnv)

	useRedisEnv, ok := os.LookupEnv("USE_REDIS")
	if !ok {
		panic("missing USE_REDIS environment")
	}
	useRedis, _ = strconv.ParseBool(useRedisEnv)

	useRSAEnv, ok := os.LookupEnv("USE_RSA_KEY")
	if !ok {
		panic("missing USE_RSA_KEY environment")
	}
	useRSAKey, _ = strconv.ParseBool(useRSAEnv)
	// ---------------------------

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

	// kafka environment
	if GlobalEnv.UseKafka {
		kafkaBrokers, ok := os.LookupEnv("KAFKA_BROKERS")
		if !ok {
			panic("kafka consumer is active, missing KAFKA_BROKERS environment")
		}
		GlobalEnv.Kafka.Brokers = strings.Split(kafkaBrokers, ",")
		GlobalEnv.Kafka.ClientID, ok = os.LookupEnv("KAFKA_CLIENT_ID")
		if !ok {
			panic("kafka consumer is active, missing KAFKA_CLIENT_ID environment")
		}
		GlobalEnv.Kafka.ConsumerGroup, ok = os.LookupEnv("KAFKA_CONSUMER_GROUP")
		if !ok {
			panic("kafka consumer is active, missing KAFKA_CONSUMER_GROUP environment")
		}
	}
}

// Exit close all connection
func (c *Config) Exit(ctx context.Context) {
	if useMongo {
		// close mongo connection
		c.MongoRead.Client().Disconnect(ctx)
		c.MongoWrite.Client().Disconnect(ctx)
	}

	if useRedis {
		// close redis connection
		c.RedisReadPool.Close()
		c.RedisWritePool.Close()
	}

	if useSQL {
		// close sql connection
		c.SQLRead.Close()
		c.SQLWrite.Close()
	}

	log.Println("\x1b[33;1mConfig: Success close all connection\x1b[0m")
}
