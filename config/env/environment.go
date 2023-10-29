package env

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/golangid/candi/candihelper"
	"github.com/joho/godotenv"
)

// Env model
type Env struct {
	RootApp, ServiceName string
	BuildNumber          string
	// Env on application
	Environment       string
	LoadConfigTimeout time.Duration

	useSQL, useMongo, useRedis, useRSAKey bool
	UseSharedListener                     bool

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
	// UsePostgresListenerWorker env
	UsePostgresListenerWorker bool
	// UseRabbitMQWorker env
	UseRabbitMQWorker bool

	DebugMode bool

	HTTPRootPath                string
	GraphQLDisableIntrospection bool

	// HTTPPort config
	HTTPPort uint16
	// GRPCPort Config
	GRPCPort uint16
	// TaskQueueDashboardPort Config
	TaskQueueDashboardPort uint16
	// TaskQueueDashboardMaxClientSubscribers Config
	TaskQueueDashboardMaxClientSubscribers int

	// BasicAuthUsername config
	BasicAuthUsername string
	// BasicAuthPassword config
	BasicAuthPassword string

	// JaegerTracingHost env
	JaegerTracingHost string
	// JaegerMaxPacketSize env
	JaegerMaxPacketSize int

	// Broker environment
	Kafka struct {
		Brokers       []string
		ClientVersion string
		ClientID      string
		ConsumerGroup string
	}
	RabbitMQ struct {
		Broker        string
		ConsumerGroup string
		ExchangeName  string
	}

	// MaxGoroutines env for goroutine semaphore
	MaxGoroutines int

	// Database environment
	DbMongoWriteHost, DbMongoReadHost string
	DbSQLWriteDSN, DbSQLReadDSN       string
	DbRedisReadDSN, DbRedisWriteDSN   string

	// CORS Environment
	CORSAllowOrigins, CORSAllowMethods, CORSAllowHeaders []string
	CORSAllowCredential                                  bool

	StartAt string
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
	var ok bool
	env.ServiceName = serviceName

	// load main .env and additional .env in app
	err := godotenv.Load(os.Getenv(candihelper.WORKDIR) + ".env")
	if err != nil {
		log.Printf("Warning: load env, %v", err)
	}

	mErrs := candihelper.NewMultiError()

	// ------------------------------------
	parseAppConfig()
	env.BuildNumber = os.Getenv("BUILD_NUMBER")

	if env.LoadConfigTimeout, err = time.ParseDuration(os.Getenv("LOAD_CONFIG_TIMEOUT")); err != nil {
		env.LoadConfigTimeout = 10 * time.Second // default value
	}

	// ------------------------------------
	isServerActive := env.UseREST || env.UseGraphQL || env.UseGRPC
	if isServerActive {
		httpPort, _ := strconv.Atoi(os.Getenv("HTTP_PORT"))
		env.HTTPPort = uint16(httpPort)
		grpcPort, _ := strconv.Atoi(os.Getenv("GRPC_PORT"))
		env.GRPCPort = uint16(grpcPort)

		env.UseSharedListener = parseBool("USE_SHARED_LISTENER")
		if env.UseSharedListener && env.HTTPPort <= 0 {
			mErrs.Append("USE_SHARED_LISTENER", errors.New("missing or invalid value for HTTP_PORT environment"))
		}
	}

	if env.UseTaskQueueWorker {
		taskQueueDashboardPort, ok := os.LookupEnv("TASK_QUEUE_DASHBOARD_PORT")
		if !ok {
			taskQueueDashboardPort = "8080"
		}
		port, err := strconv.Atoi(taskQueueDashboardPort)
		if err != nil {
			mErrs.Append("TASK_QUEUE_DASHBOARD_PORT", errors.New("TASK_QUEUE_DASHBOARD_PORT environment must in integer format"))
		}
		env.TaskQueueDashboardPort = uint16(port)
		env.TaskQueueDashboardMaxClientSubscribers, _ = strconv.Atoi(os.Getenv("TASK_QUEUE_DASHBOARD_MAX_CLIENT"))
		if env.TaskQueueDashboardPort <= 0 || env.TaskQueueDashboardMaxClientSubscribers > 10 {
			env.TaskQueueDashboardMaxClientSubscribers = 10 // default
		}
	}

	// ------------------------------------
	env.Environment = os.Getenv("ENVIRONMENT")
	env.DebugMode, err = strconv.ParseBool(os.Getenv("DEBUG_MODE"))
	if err != nil {
		env.DebugMode = true
	}

	env.GraphQLDisableIntrospection = parseBool("GRAPHQL_DISABLE_INTROSPECTION")
	env.HTTPRootPath = os.Getenv("HTTP_ROOT_PATH")

	env.BasicAuthUsername, ok = os.LookupEnv("BASIC_AUTH_USERNAME")
	if !ok {
		mErrs.Append("BASIC_AUTH_USERNAME", errors.New("missing BASIC_AUTH_USERNAME environment"))
	}
	env.BasicAuthPassword, ok = os.LookupEnv("BASIC_AUTH_PASS")
	if !ok {
		mErrs.Append("BASIC_AUTH_PASS", errors.New("missing BASIC_AUTH_PASS environment"))
	}

	env.JaegerTracingHost = os.Getenv("JAEGER_TRACING_HOST")
	jaegerMaxpacketSize, err := strconv.Atoi(os.Getenv("JAEGER_MAX_PACKET_SIZE"))
	if err != nil || jaegerMaxpacketSize < 0 {
		jaegerMaxpacketSize = 65000 // default max packet size of UDP
	}
	env.JaegerMaxPacketSize = int(jaegerMaxpacketSize) * int(candihelper.Byte)

	// kafka environment
	parseBrokerEnv(mErrs)

	maxGoroutines, err := strconv.Atoi(os.Getenv("MAX_GOROUTINES"))
	if err != nil || maxGoroutines <= 0 {
		maxGoroutines = 10
	}
	env.MaxGoroutines = maxGoroutines

	// Parse database environment
	parseDatabaseEnv()

	// Parse CORS environment
	parseCorsEnv()

	env.StartAt = time.Now().Format(time.RFC3339)

	if mErrs.HasError() {
		panic("Basic environment error: \n" + mErrs.Error())
	}
}

func parseAppConfig() {

	useREST, ok := os.LookupEnv("USE_REST")
	if !ok {
		flag.BoolVar(&env.UseREST, "USE_REST", false, "USE REST")
	} else {
		env.UseREST, _ = strconv.ParseBool(useREST)
	}

	useGraphQL, ok := os.LookupEnv("USE_GRAPHQL")
	if !ok {
		flag.BoolVar(&env.UseGraphQL, "USE_GRAPHQL", false, "USE GRAPHQL")
	} else {
		env.UseGraphQL, _ = strconv.ParseBool(useGraphQL)
	}

	useGRPC, ok := os.LookupEnv("USE_GRPC")
	if !ok {
		flag.BoolVar(&env.UseGRPC, "USE_GRPC", false, "USE GRPC")
	} else {
		env.UseGRPC, _ = strconv.ParseBool(useGRPC)
	}

	useKafkaConsumer, ok := os.LookupEnv("USE_KAFKA_CONSUMER")
	if !ok {
		flag.BoolVar(&env.UseKafkaConsumer, "USE_KAFKA_CONSUMER", false, "USE KAFKA CONSUMER")
	} else {
		env.UseKafkaConsumer, _ = strconv.ParseBool(useKafkaConsumer)
	}

	useCronScheduler, ok := os.LookupEnv("USE_CRON_SCHEDULER")
	if !ok {
		flag.BoolVar(&env.UseCronScheduler, "USE_CRON_SCHEDULER", false, "USE CRON SCHEDULER")
	} else {
		env.UseCronScheduler, _ = strconv.ParseBool(useCronScheduler)
	}

	useRedisSubs, ok := os.LookupEnv("USE_REDIS_SUBSCRIBER")
	if !ok {
		flag.BoolVar(&env.UseRedisSubscriber, "USE_REDIS_SUBSCRIBER", false, "USE REDIS SUBSCRIBER")
	} else {
		env.UseRedisSubscriber, _ = strconv.ParseBool(useRedisSubs)
	}

	useTaskQueue, ok := os.LookupEnv("USE_TASK_QUEUE_WORKER")
	if !ok {
		flag.BoolVar(&env.UseTaskQueueWorker, "USE_TASK_QUEUE_WORKER", false, "USE TASK QUEUE WORKER")
	} else {
		env.UseTaskQueueWorker, _ = strconv.ParseBool(useTaskQueue)
	}
	usePostgresListener, ok := os.LookupEnv("USE_POSTGRES_LISTENER_WORKER")
	if !ok {
		flag.BoolVar(&env.UsePostgresListenerWorker, "USE_POSTGRES_LISTENER_WORKER", false, "USE POSTGRES LISTENER WORKER")
	} else {
		env.UsePostgresListenerWorker, _ = strconv.ParseBool(usePostgresListener)
	}
	useRabbitMQWorker, ok := os.LookupEnv("USE_RABBITMQ_CONSUMER")
	if !ok {
		flag.BoolVar(&env.UseRabbitMQWorker, "USE_RABBITMQ_CONSUMER", false, "USE RABBIT MQ CONSUMER")
	} else {
		env.UseRabbitMQWorker, _ = strconv.ParseBool(useRabbitMQWorker)
	}

	flag.Usage = func() {
		fmt.Println("	-USE_REST :=> Activate REST Server")
		fmt.Println("	-USE_GRPC :=> Activate GRPC Server")
		fmt.Println("	-USE_GRAPHQL :=> Activate GraphQL Server")
		fmt.Println("	-USE_KAFKA_CONSUMER :=> Activate Kafka Consumer Worker")
		fmt.Println("	-USE_CRON_SCHEDULER :=> Activate Cron Scheduler Worker")
		fmt.Println("	-USE_REDIS_SUBSCRIBER :=> Activate Redis Subscriber Worker")
		fmt.Println("	-USE_TASK_QUEUE_WORKER :=> Activate Task Queue Worker")
		fmt.Println("	-USE_POSTGRES_LISTENER_WORKER :=> Activate Postgres Event Worker")
		fmt.Println("	-USE_RABBITMQ_CONSUMER :=> Activate Rabbit MQ Consumer")
	}
	flag.Parse()
}

func parseBrokerEnv(mErrs candihelper.MultiError) {
	kafkaBrokerEnv := os.Getenv("KAFKA_BROKERS")
	env.Kafka.Brokers = strings.Split(kafkaBrokerEnv, ",") // optional
	env.Kafka.ClientID = os.Getenv("KAFKA_CLIENT_ID")      // optional
	env.Kafka.ClientVersion = os.Getenv("KAFKA_CLIENT_VERSION")
	if env.UseKafkaConsumer {
		if kafkaBrokerEnv == "" {
			mErrs.Append("KAFKA_BROKERS", errors.New("kafka consumer is active, missing KAFKA_BROKERS environment"))
		}

		var ok bool
		env.Kafka.ConsumerGroup, ok = os.LookupEnv("KAFKA_CONSUMER_GROUP")
		if !ok {
			mErrs.Append("KAFKA_CONSUMER_GROUP", errors.New("kafka consumer is active, missing KAFKA_CONSUMER_GROUP environment"))
		}
	}
	env.RabbitMQ.Broker = os.Getenv("RABBITMQ_BROKER")
	env.RabbitMQ.ConsumerGroup = os.Getenv("RABBITMQ_CONSUMER_GROUP")
	env.RabbitMQ.ExchangeName = os.Getenv("RABBITMQ_EXCHANGE_NAME")
}

func parseDatabaseEnv() {
	env.DbMongoWriteHost = os.Getenv("MONGODB_HOST_WRITE")
	env.DbMongoReadHost = os.Getenv("MONGODB_HOST_READ")

	env.DbSQLReadDSN = os.Getenv("SQL_DB_READ_DSN")
	env.DbSQLWriteDSN = os.Getenv("SQL_DB_WRITE_DSN")

	env.DbRedisReadDSN = os.Getenv("REDIS_READ_DSN")
	env.DbRedisWriteDSN = os.Getenv("REDIS_WRITE_DSN")
}

func parseCorsEnv() {
	CORSAllowOrigins := os.Getenv("CORS_ALLOW_ORIGINS")
	if CORSAllowOrigins == "" {
		env.CORSAllowOrigins = []string{"*"}
	} else {
		env.CORSAllowOrigins = strings.Split(CORSAllowOrigins, ",")
	}
	CORSAllowMethods := os.Getenv("CORS_ALLOW_METHODS")
	if CORSAllowMethods == "" {
		env.CORSAllowMethods = []string{
			http.MethodGet,
			http.MethodHead,
			http.MethodPut,
			http.MethodPatch,
			http.MethodPost,
			http.MethodDelete,
		}
	} else {
		env.CORSAllowMethods = strings.Split(CORSAllowMethods, ",")
	}
	CORSAllowHeaders := os.Getenv("CORS_ALLOW_HEADERS")
	if CORSAllowHeaders != "" {
		env.CORSAllowHeaders = strings.Split(CORSAllowHeaders, ",")
	}
	env.CORSAllowCredential, _ = strconv.ParseBool(os.Getenv("CORS_ALLOW_CREDENTIAL"))
}

func parseBool(envName string) bool {
	b, _ := strconv.ParseBool(os.Getenv(envName))
	return b
}
