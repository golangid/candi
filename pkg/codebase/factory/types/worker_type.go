package types

import "context"

// Worker is the type returned by a classifier worker (kafka, redis subscriber, rabbitmq, scheduler)
type Worker int

const (
	// Kafka worker
	Kafka Worker = iota
	// RedisSubscriber worker
	RedisSubscriber
	// RabbitMQ worker
	RabbitMQ
	// Scheduler worker
	Scheduler
	// TaskQueue worker
	TaskQueue
)

// WorkerHandlerFunc types
type WorkerHandlerFunc func(context.Context, []byte) error
