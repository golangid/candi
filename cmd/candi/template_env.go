package main

const envTemplate = `# Basic Candi env configuration
ENVIRONMENT=development #development,staging,production
DEBUG_MODE=true
LOAD_CONFIG_TIMEOUT=20s

# Application Service Handlers
## Server
USE_REST={{.RestHandler}}
USE_GRPC={{.GRPCHandler}}
USE_GRAPHQL={{.GraphQLHandler}}
## Worker
USE_KAFKA_CONSUMER={{.KafkaHandler}} # event driven handler
USE_CRON_SCHEDULER={{.SchedulerHandler}} # static scheduler
USE_REDIS_SUBSCRIBER={{.RedisSubsHandler}} # dynamic scheduler
USE_TASK_QUEUE_WORKER={{.TaskQueueHandler}}
USE_POSTGRES_LISTENER_WORKER={{.PostgresListenerHandler}}
USE_RABBITMQ_CONSUMER={{.RabbitMQHandler}} # event driven handler and dynamic scheduler

# use shared listener setup shared port to http & grpc listener (if true, use HTTP_PORT value)
USE_SHARED_LISTENER=false
HTTP_PORT=8000
GRPC_PORT=8002

TASK_QUEUE_DASHBOARD_PORT=8080
TASK_QUEUE_DASHBOARD_MAX_CLIENT=5

GRAPHQL_DISABLE_INTROSPECTION=false
HTTP_ROOT_PATH=""

BASIC_AUTH_USERNAME=user
BASIC_AUTH_PASS=pass

MONGODB_HOST_WRITE=mongodb://user:pass@localhost:27017/{{.ServiceName}}?authSource=admin
MONGODB_HOST_READ=mongodb://user:pass@localhost:27017/{{.ServiceName}}?authSource=admin

SQL_DB_READ_DSN={{ if .SQLDeps }}{{.SQLDriver}}://` +
	"{{if eq .SQLDriver \"postgres\"}}user:pass@localhost:5432/db_name?sslmode=disable&TimeZone=Asia/Jakarta{{else if eq .SQLDriver \"mysql\"}}" +
	"root:pass@tcp(127.0.0.1:3306)/db_name?charset=utf8&parseTime=True{{else if eq .SQLDriver \"sqlite3\"}}database.db{{end}}" +
	`{{ end }}
SQL_DB_WRITE_DSN={{ if .SQLDeps }}{{.SQLDriver}}://` +
	"{{if eq .SQLDriver \"postgres\"}}user:pass@localhost:5432/db_name?sslmode=disable&TimeZone=Asia/Jakarta{{else if eq .SQLDriver \"mysql\"}}" +
	"root:pass@tcp(127.0.0.1:3306)/db_name?charset=utf8&parseTime=True{{else if eq .SQLDriver \"sqlite3\"}}database.db{{end}}" +
	`{{ end }}
{{if .ArangoDeps}}
ARANGODB_HOST_WRITE=http://user:pass@localhost:8529/{{.ServiceName}}
ARANGODB_HOST_READ=http://user:pass@localhost:8529/{{.ServiceName}}
{{end}}
REDIS_READ_DSN=redis://:pass@localhost:6379/0
REDIS_WRITE_DSN=redis://:pass@localhost:6379/0

KAFKA_BROKERS=localhost:9092 # if multiple broker, separate by comma with no space
KAFKA_CLIENT_VERSION=2.0.0
KAFKA_CLIENT_ID={{.ServiceName}}
KAFKA_CONSUMER_GROUP={{.ServiceName}}

{{if not .RabbitMQHandler}}# {{end}}RABBITMQ_BROKER=amqp://guest:guest@localhost:5672/test
{{if not .RabbitMQHandler}}# {{end}}RABBITMQ_CONSUMER_GROUP={{.ServiceName}}
{{if not .RabbitMQHandler}}# {{end}}RABBITMQ_EXCHANGE_NAME=delayed

JAEGER_TRACING_HOST=127.0.0.1:5775
JAEGER_MAX_PACKET_SIZE=65000 # in bytes

MAX_GOROUTINES=10

# Additional env for your service

`
