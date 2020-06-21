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
	// UseKafkaConsumer env
	UseKafkaConsumer bool

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

	useKafkaConsumer, ok := os.LookupEnv("USE_KAFKA_CONSUMER")
	if !ok {
		panic("missing USE_KAFKA_CONSUMER environment")
	}
	env.UseKafkaConsumer, _ = strconv.ParseBool(useKafkaConsumer)

	// ------------------------------------

	if env.UseHTTP {
		if httpPort, ok := os.LookupEnv("HTTP_PORT"); !ok {
			panic("missing HTTP_PORT environment")
		} else {
			port, err := strconv.Atoi(httpPort)
			if err != nil {
				panic("HTTP_PORT environment must in integer format")
			}
			env.HTTPPort = uint16(port)
		}
	}

	if env.UseGRPC {
		if grpcPort, ok := os.LookupEnv("GRPC_PORT"); !ok {
			panic("missing GRPC_PORT environment")
		} else {
			port, err := strconv.Atoi(grpcPort)
			if err != nil {
				panic("GRPC_PORT environment must in integer format")
			}
			env.GRPCPort = uint16(port)
		}
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

	kafkaBrokerEnv := os.Getenv("KAFKA_BROKERS")
	env.Kafka.Brokers = strings.Split(kafkaBrokerEnv, ",") // optional
	env.Kafka.ClientID = os.Getenv("KAFKA_CLIENT_ID")      // optional

	// kafka environment
	if env.UseKafkaConsumer {
		if kafkaBrokerEnv == "" {
			panic("kafka consumer is active, missing KAFKA_BROKERS environment")
		}

		env.Kafka.ConsumerGroup, ok = os.LookupEnv("KAFKA_CONSUMER_GROUP")
		if !ok {
			panic("kafka consumer is active, missing KAFKA_CONSUMER_GROUP environment")
		}
	}

}
