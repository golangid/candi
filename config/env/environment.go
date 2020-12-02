package env

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	"pkg.agungdwiprasetyo.com/candi/candihelper"
)

// Env model
type Env struct {
	RootApp, ServiceName string

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

	// Env on application
	Environment string

	IsProduction, DebugMode bool

	// HTTPPort config
	HTTPPort uint16
	// GRPCPort Config
	GRPCPort uint16
	// TaskQueueDashboardPort Config
	TaskQueueDashboardPort uint16

	// BasicAuthUsername config
	BasicAuthUsername string
	// BasicAuthPassword config
	BasicAuthPassword string

	// JaegerTracingHost env
	JaegerTracingHost string

	// Broker environment
	Kafka struct {
		Brokers       []string
		ClientVersion string
		ClientID      string
		ConsumerGroup string
	}

	// MaxGoroutines env for goroutine semaphore
	MaxGoroutines int

	// Database environment
	DbMongoWriteHost, DbMongoReadHost, DbMongoDatabaseName                    string
	DbSQLWriteHost, DbSQLWriteUser, DbSQLWritePass                            string
	DbSQLReadHost, DbSQLReadUser, DbSQLReadPass                               string
	DbSQLDatabaseName, DbSQLDriver                                            string
	DbRedisReadHost, DbRedisReadPort, DbRedisReadAuth, DbRedisReadDBIndex     string
	DbRedisWriteHost, DbRedisWritePort, DbRedisWriteAuth, DbRedisWriteDBIndex string
	DbRedisReadTLS, DbRedisWriteTLS                                           bool
}

var env Env

// BaseEnv get global basic environment
func BaseEnv() Env {
	return env
}

// SetEnv set env for mocking data env
func SetEnv(newEnv Env) {
	env = newEnv
}

// Load environment
func Load(serviceName string) {
	env.ServiceName = serviceName

	// load main .env and additional .env in app
	err := godotenv.Load(os.Getenv(candihelper.WORKDIR) + ".env")
	if err != nil {
		panic(fmt.Errorf("Load env: %v", err))
	}

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

	useTaskQueue, ok := os.LookupEnv("USE_TASK_QUEUE_WORKER")
	if !ok {
		panic("missing USE_TASK_QUEUE_WORKER environment")
	}
	env.UseTaskQueueWorker, _ = strconv.ParseBool(useTaskQueue)

	// ------------------------------------
	if env.UseREST || env.UseGraphQL {
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

	if env.UseTaskQueueWorker {
		taskQueueDashboardPort, ok := os.LookupEnv("TASK_QUEUE_DASHBOARD_PORT")
		if !ok {
			taskQueueDashboardPort = "8080"
		}
		port, err := strconv.Atoi(taskQueueDashboardPort)
		if err != nil {
			panic("TASK_QUEUE_DASHBOARD_PORT environment must in integer format")
		}
		env.TaskQueueDashboardPort = uint16(port)
	}

	// ------------------------------------
	env.Environment = os.Getenv("ENVIRONMENT")
	env.IsProduction = strings.ToLower(env.Environment) == "production"
	env.DebugMode, err = strconv.ParseBool(os.Getenv("DEBUG_MODE"))
	if err != nil {
		env.DebugMode = true
	}
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

	// kafka environment
	parseBrokerEnv()

	env.GraphQLSchemaDir, ok = os.LookupEnv("GRAPHQL_SCHEMA_DIR")
	if env.UseGraphQL && !ok {
		panic("GRAPHQL is active, missing GRAPHQL_SCHEMA_DIR environment")
	}
	env.GraphQLSchemaDir = os.Getenv(candihelper.WORKDIR) + env.GraphQLSchemaDir

	env.JSONSchemaDir, ok = os.LookupEnv("JSON_SCHEMA_DIR")
	if !ok {
		panic("missing JSON_SCHEMA_DIR environment")
	}
	env.JSONSchemaDir = os.Getenv(candihelper.WORKDIR) + env.JSONSchemaDir

	maxGoroutines, err := strconv.Atoi(os.Getenv("MAX_GOROUTINES"))
	if err != nil || maxGoroutines <= 0 {
		maxGoroutines = 4096
	}
	env.MaxGoroutines = maxGoroutines

	// Parse database environment
	parseDatabaseEnv()
}

func parseBrokerEnv() {
	kafkaBrokerEnv := os.Getenv("KAFKA_BROKERS")
	env.Kafka.Brokers = strings.Split(kafkaBrokerEnv, ",") // optional
	env.Kafka.ClientID = os.Getenv("KAFKA_CLIENT_ID")      // optional
	env.Kafka.ClientVersion = os.Getenv("KAFKA_CLIENT_VERSION")
	if env.UseKafkaConsumer {
		if kafkaBrokerEnv == "" {
			panic("kafka consumer is active, missing KAFKA_BROKERS environment")
		}

		var ok bool
		env.Kafka.ConsumerGroup, ok = os.LookupEnv("KAFKA_CONSUMER_GROUP")
		if !ok {
			panic("kafka consumer is active, missing KAFKA_CONSUMER_GROUP environment")
		}
	}
}

func parseDatabaseEnv() {
	env.DbMongoWriteHost = os.Getenv("MONGODB_HOST_WRITE")
	env.DbMongoReadHost = os.Getenv("MONGODB_HOST_READ")
	env.DbMongoDatabaseName = os.Getenv("MONGODB_DATABASE_NAME")

	env.DbSQLDriver = os.Getenv("SQL_DRIVER_NAME")
	env.DbSQLReadHost = os.Getenv("SQL_DB_READ_HOST")
	env.DbSQLReadUser = os.Getenv("SQL_DB_READ_USER")
	env.DbSQLReadPass = os.Getenv("SQL_DB_READ_PASSWORD")
	env.DbSQLWriteHost = os.Getenv("SQL_DB_WRITE_HOST")
	env.DbSQLWriteUser = os.Getenv("SQL_DB_WRITE_USER")
	env.DbSQLWritePass = os.Getenv("SQL_DB_WRITE_PASSWORD")
	env.DbSQLDatabaseName = os.Getenv("SQL_DATABASE_NAME")

	env.DbRedisReadHost = os.Getenv("REDIS_READ_HOST")
	env.DbRedisReadPort = os.Getenv("REDIS_READ_PORT")
	env.DbRedisReadAuth = os.Getenv("REDIS_READ_AUTH")
	env.DbRedisReadTLS, _ = strconv.ParseBool(os.Getenv("REDIS_READ_TLS"))
	env.DbRedisReadDBIndex = os.Getenv("REDIS_READ_DB_INDEX")
	env.DbRedisWriteHost = os.Getenv("REDIS_WRITE_HOST")
	env.DbRedisWritePort = os.Getenv("REDIS_WRITE_PORT")
	env.DbRedisWriteAuth = os.Getenv("REDIS_WRITE_AUTH")
	env.DbRedisWriteTLS, _ = strconv.ParseBool(os.Getenv("REDIS_WRITE_TLS"))
	env.DbRedisWriteDBIndex = os.Getenv("REDIS_WRITE_DB_INDEX")
}
