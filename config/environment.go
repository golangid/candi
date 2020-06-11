package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Env model
type Env struct {
	RootApp string

	useSQL, useMongo, useRedis, useRSAKey bool

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

	// CacheExpired config
	CacheExpired time.Duration

	Kafka struct {
		Brokers       []string
		ClientID      string
		ConsumerGroup string
	}
}

func loadBaseEnv(serviceLocation string, targetEnv *Env) {

	// load main .env and additional .env in app
	if err := godotenv.Load(serviceLocation + ".env"); err != nil {
		panic(fmt.Errorf("Load env: %v", err))
	}

	rootApp, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	os.Setenv("APP_PATH", rootApp)
	env.RootApp = rootApp

	// ------------------------------------
	useHTTP, ok := os.LookupEnv("USE_HTTP")
	if !ok {
		panic("missing USE_HTTP environment")
	}
	env.UseHTTP, _ = strconv.ParseBool(useHTTP)

	useGraphQL, ok := os.LookupEnv("USE_GRAPHQL")
	if !ok {
		panic("missing USE_GRAPHQL environment")
	}
	env.UseGraphQL, _ = strconv.ParseBool(useGraphQL)
	if env.UseGraphQL && !env.UseHTTP {
		panic("GraphQL required http server")
	}

	useGRPC, ok := os.LookupEnv("USE_GRPC")
	if !ok {
		panic("missing USE_GRPC environment")
	}
	env.UseGRPC, _ = strconv.ParseBool(useGRPC)

	useKafka, ok := os.LookupEnv("USE_KAFKA")
	if !ok {
		panic("missing USE_KAFKA environment")
	}
	env.UseKafka, _ = strconv.ParseBool(useKafka)

	// ---------------------------
	useMongoEnv, ok := os.LookupEnv("USE_MONGO")
	if !ok {
		panic("missing USE_MONGO environment")
	}
	env.useMongo, _ = strconv.ParseBool(useMongoEnv)

	useSQLEnv, ok := os.LookupEnv("USE_SQL")
	if !ok {
		panic("missing USE_SQL environment")
	}
	env.useSQL, _ = strconv.ParseBool(useSQLEnv)

	useRedisEnv, ok := os.LookupEnv("USE_REDIS")
	if !ok {
		panic("missing USE_REDIS environment")
	}
	env.useRedis, _ = strconv.ParseBool(useRedisEnv)

	useRSAEnv, ok := os.LookupEnv("USE_RSA_KEY")
	if !ok {
		panic("missing USE_RSA_KEY environment")
	}
	env.useRSAKey, _ = strconv.ParseBool(useRSAEnv)
	// ---------------------------

	// ------------------------------------

	if port, err := strconv.Atoi(os.Getenv("HTTP_PORT")); err != nil {
		panic("missing HTTP_PORT environment")
	} else {
		env.HTTPPort = uint16(port)
	}

	if port, err := strconv.Atoi(os.Getenv("GRPC_PORT")); err != nil {
		panic("missing GRPC_PORT environment")
	} else {
		env.GRPCPort = uint16(port)
	}

	// ------------------------------------
	env.BasicAuthUsername, ok = os.LookupEnv("BASIC_AUTH_USERNAME")
	if !ok {
		panic("missing BASIC_AUTH_USERNAME environment")
	}
	env.BasicAuthPassword, ok = os.LookupEnv("BASIC_AUTH_PASS")
	if !ok {
		panic("missing BASIC_AUTH_PASS environment")
	}

	// kafka environment
	if env.UseKafka {
		kafkaBrokers, ok := os.LookupEnv("KAFKA_BROKERS")
		if !ok {
			panic("kafka consumer is active, missing KAFKA_BROKERS environment")
		}
		env.Kafka.Brokers = strings.Split(kafkaBrokers, ",")
		env.Kafka.ClientID, ok = os.LookupEnv("KAFKA_CLIENT_ID")
		if !ok {
			panic("kafka consumer is active, missing KAFKA_CLIENT_ID environment")
		}
		env.Kafka.ConsumerGroup, ok = os.LookupEnv("KAFKA_CONSUMER_GROUP")
		if !ok {
			panic("kafka consumer is active, missing KAFKA_CONSUMER_GROUP environment")
		}
	}

}
