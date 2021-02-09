package main

const envTemplate = `# Basic env configuration
ENVIRONMENT=development #development,staging,production
DEBUG_MODE=true
NO_AUTH=true

# Service Handlers
## Server
USE_REST={{.RestHandler}}
USE_GRPC={{.GRPCHandler}}
USE_GRAPHQL={{.GraphQLHandler}}
## Worker
USE_KAFKA_CONSUMER={{.KafkaHandler}}
USE_CRON_SCHEDULER={{.SchedulerHandler}}
USE_REDIS_SUBSCRIBER={{.RedisSubsHandler}}
USE_TASK_QUEUE_WORKER={{.TaskQueueHandler}}

HTTP_PORT=8000
GRPC_PORT=8002 # uncomment this env if separate port listener http & grpc server, comment this env if use shared listener from http & grpc in same port (use HTTP_PORT)

TASK_QUEUE_DASHBOARD_PORT=8080
TASK_QUEUE_DASHBOARD_MAX_CLIENT=5

# use consul for distributed lock if run in multiple instance
USE_CONSUL=false
CONSUL_AGENT_HOST=127.0.0.1:8500
CONSUL_MAX_JOB_REBALANCE=10 # if worker execute total job in env config, rebalance worker to another active intance

BASIC_AUTH_USERNAME=user
BASIC_AUTH_PASS=pass

MONGODB_HOST_WRITE=mongodb://user:pass@localhost:27017
MONGODB_HOST_READ=mongodb://user:pass@localhost:27017
MONGODB_DATABASE_NAME={{.ServiceName}}

SQL_DB_READ_DSN={{ if .SQLDeps }}{{.SQLDriver}}{{ else }}sql{{ end }}://user:pass@localhost:5432/db_name?sslmode=disable
SQL_DB_WRITE_DSN={{ if .SQLDeps }}{{.SQLDriver}}{{ else }}sql{{ end }}://user:pass@localhost:5432/db_name?sslmode=disable

REDIS_READ_DSN=redis://:pass@localhost:6379/0
REDIS_WRITE_DSN=redis://:pass@localhost:6379/0

KAFKA_BROKERS=localhost:9092
KAFKA_CLIENT_VERSION=2.0.0
KAFKA_CLIENT_ID={{.ServiceName}}
KAFKA_CONSUMER_GROUP={{.ServiceName}}

JAEGER_TRACING_HOST=127.0.0.1:5775
GRAPHQL_SCHEMA_DIR="api/graphql"
JSON_SCHEMA_DIR="api/jsonschema"

MAX_GOROUTINES=100

# Additional env
`
