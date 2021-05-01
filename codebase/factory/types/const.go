package types

// Server is the type returned by a classifier server (REST, gRPC, GraphQL)
type Server string

// Worker is the type returned by a classifier worker (kafka, redis subscriber, rabbitmq, scheduler, task queue)
type Worker string

const (
	// REST server
	REST Server = "rest"
	// GRPC server
	GRPC Server = "grpc"
	// GraphQL server
	GraphQL Server = "graphql"

	// Kafka worker
	Kafka Worker = "kafka"
	// RedisSubscriber worker
	RedisSubscriber Worker = "redis_subscriber"
	// RabbitMQ worker
	RabbitMQ Worker = "rabbit_mq"
	// Scheduler worker
	Scheduler Worker = "scheduler"
	// TaskQueue worker
	TaskQueue Worker = "task_queue"
	// PostgresListener worker
	PostgresListener Worker = "postgres_listener"
)
