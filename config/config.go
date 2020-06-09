package config

import (
	"context"
	"crypto/rsa"
	"database/sql"
	"fmt"
	"log"

	"agungdwiprasetyo.com/backend-microservices/config/broker"
	"agungdwiprasetyo.com/backend-microservices/config/database"
	"agungdwiprasetyo.com/backend-microservices/config/key"
	"github.com/Shopify/sarama"
	"github.com/gomodule/redigo/redis"
	"go.mongodb.org/mongo-driver/mongo"
)

var env Env

// Config app
type Config struct {
	MongoRead, MongoWrite         *mongo.Database
	SQLRead, SQLWrite             *sql.DB
	RedisReadPool, RedisWritePool *redis.Pool
	PrivateKey                    *rsa.PrivateKey
	PublicKey                     *rsa.PublicKey
	KafkaConsumerConfig           *sarama.Config
}

// Init app config
func Init(ctx context.Context, rootApp string) *Config {
	loadBaseEnv(rootApp, &env)

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
		cfg.MongoRead, cfg.MongoWrite = database.InitMongoDB(ctx, env.useMongo)
		cfg.SQLRead, cfg.SQLWrite = database.InitSQLDatabase(ctx, env.useSQL)
		cfg.RedisReadPool, cfg.RedisWritePool = database.InitRedis(env.useRedis)
		cfg.PrivateKey, cfg.PublicKey = key.LoadRSAKey(env.useRSAKey)
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

// BaseEnv get global basic environment
func BaseEnv() Env {
	return env
}

// Exit close all connection
func (c *Config) Exit(ctx context.Context) {
	if env.useMongo {
		// close mongo connection
		c.MongoRead.Client().Disconnect(ctx)
		c.MongoWrite.Client().Disconnect(ctx)
	}

	if env.useRedis {
		// close redis connection
		c.RedisReadPool.Close()
		c.RedisWritePool.Close()
	}

	if env.useSQL {
		// close sql connection
		c.SQLRead.Close()
		c.SQLWrite.Close()
	}

	log.Println("\x1b[33;1mConfig: Success close all connection\x1b[0m")
}
