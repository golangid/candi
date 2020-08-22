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

	// UseREST env
	UseREST bool
	// UseGraphQL env
	UseGraphQL bool
	// UseGRPC env
	UseGRPC bool
	// UseKafkaConsumer env
	UseKafkaConsumer bool
	// UseCronScheduler env
	UseCronScheduler bool
	// UseRedisSubscriber env
	UseRedisSubscriber bool
	// UseTaskQueueWorker env
	UseTaskQueueWorker bool

	// GraphQLSchemaDir env
	GraphQLSchemaDir string
	// JSONSchemaDir env
	JSONSchemaDir string

	// Development env checking, this env for debug purpose
	Development string

	// Env on application
	Environment string

	// RESTPort config
	RESTPort uint16
	// GraphQLPort config
	GraphQLPort uint16
	// GRPCPort Config
	GRPCPort uint16

	// BasicAuthUsername config
	BasicAuthUsername string
	// BasicAuthPassword config
	BasicAuthPassword string

	// CacheExpired config
	CacheExpired time.Duration

	// JaegerTracingHost env
	JaegerTracingHost string

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
	useREST, ok := os.LookupEnv("USE_REST")
	if !ok {
		panic("missing USE_REST environment")
	}
	env.UseREST, _ = strconv.ParseBool(useREST)

	useGraphQL, ok := os.LookupEnv("USE_GRAPHQL")
	if !ok {
		panic("missing USE_GRAPHQL environment")
	}
	env.UseGraphQL, _ = strconv.ParseBool(useGraphQL)

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

	useCronScheduler, ok := os.LookupEnv("USE_CRON_SCHEDULER")
	if !ok {
		panic("missing USE_CRON_SCHEDULER environment")
	}
	env.UseCronScheduler, _ = strconv.ParseBool(useCronScheduler)

	useRedisSubs, ok := os.LookupEnv("USE_REDIS_SUBSCRIBER")
	if !ok {
		panic("missing USE_REDIS_SUBSCRIBER environment")
	}
	env.UseRedisSubscriber, _ = strconv.ParseBool(useRedisSubs)

	useTaskQueueWorker, ok := os.LookupEnv("USE_TASK_QUEUE_WORKER")
	if !ok {
		panic("missing USE_TASK_QUEUE_WORKER environment")
	}
	env.UseTaskQueueWorker, _ = strconv.ParseBool(useTaskQueueWorker)

	// ------------------------------------

	if env.UseREST {
		if restPort, ok := os.LookupEnv("REST_HTTP_PORT"); !ok {
			panic("missing REST_HTTP_PORT environment")
		} else {
			port, err := strconv.Atoi(restPort)
			if err != nil {
				panic("REST_HTTP_PORT environment must in integer format")
			}
			env.RESTPort = uint16(port)
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

	if env.UseGraphQL {
		if graphqlPort, ok := os.LookupEnv("GRAPHQL_HTTP_PORT"); !ok {
			panic("missing GRAPHQL_HTTP_PORT environment")
		} else {
			port, err := strconv.Atoi(graphqlPort)
			if err != nil {
				panic("GRAPHQL_HTTP_PORT environment must in integer format")
			}
			env.GraphQLPort = uint16(port)
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

	env.JaegerTracingHost, ok = os.LookupEnv("JAEGER_TRACING_HOST")
	if !ok {
		panic("missing JAEGER_TRACING_HOST environment")
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

	env.GraphQLSchemaDir, ok = os.LookupEnv("GRAPHQL_SCHEMA_DIR")
	if env.UseGraphQL && !ok {
		panic("GRAPHQL is active, missing GRAPHQL_SCHEMA_DIR environment")
	}

	env.JSONSchemaDir, ok = os.LookupEnv("JSON_SCHEMA_DIR")
	if !ok {
		panic("missing JSON_SCHEMA_DIR environment")
	}
}
