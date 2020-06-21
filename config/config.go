package config

import (
	"context"
	"fmt"
	"log"
	"time"

	"agungdwiprasetyo.com/backend-microservices/config/broker"
	"agungdwiprasetyo.com/backend-microservices/config/database"
	"agungdwiprasetyo.com/backend-microservices/pkg/codebase/interfaces"
	"github.com/Shopify/sarama"
)

var env Env

// Config app
type Config struct {
	MongoDB     interfaces.MongoDatabase
	SQLDB       interfaces.SQLDatabase
	RedisPool   interfaces.RedisPool
	Key         interfaces.Key
	KafkaConfig *sarama.Config
}

// Init app config
func Init(rootApp string) *Config {
	// set timeout for init configuration
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

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
		cfg.MongoDB = database.InitMongoDB(ctx, env.useMongo)
		cfg.SQLDB = database.InitSQLDatabase(env.useSQL)
		cfg.RedisPool = database.InitRedis(env.useRedis)
		cfg.KafkaConfig = broker.InitKafkaConfig(env.UseKafkaConsumer, env.Kafka.ClientID)

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
func (c *Config) Exit() {
	// set timeout for close all configuration
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	if env.useMongo {
		// close mongo connection
		c.MongoDB.Disconnect(ctx)
	}

	if env.useRedis {
		// close redis connection
		c.RedisPool.Disconnect()
	}

	if env.useSQL {
		// close sql connection
		c.SQLDB.Disconnect()
	}

	log.Println("\x1b[32;1mConfig: Success close all connection\x1b[0m")
}
