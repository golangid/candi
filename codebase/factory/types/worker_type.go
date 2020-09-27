package types

import "context"

// Worker is the type returned by a classifier worker (kafka, redis subscriber, rabbitmq, scheduler, task queue)
type Worker string

const (
	// Kafka worker
	Kafka Worker = "kafka"
	// RedisSubscriber worker
	RedisSubscriber = "redis_subscriber"
	// RabbitMQ worker
	RabbitMQ = "rabbit_mq"
	// Scheduler worker
	Scheduler = "scheduler"
	// TaskQueue worker
	TaskQueue = "task_queue"
)

// WorkerHandlerFunc types
type WorkerHandlerFunc func(ctx context.Context, message []byte) error

// WorkerErrorHandler types
type WorkerErrorHandler func(ctx context.Context, workerType Worker, workerName string, message []byte, err error)

// WorkerHandlerGroup group of worker handlers by pattern string
type WorkerHandlerGroup struct {
	Handlers []struct {
		Pattern      string
		HandlerFunc  WorkerHandlerFunc
		ErrorHandler []WorkerErrorHandler
	}
}

// Add method from WorkerHandlerGroup, pattern can contains unique topic name, key, and task name
func (m *WorkerHandlerGroup) Add(pattern string, handlerFunc WorkerHandlerFunc, errHandlers ...WorkerErrorHandler) {
	m.Handlers = append(m.Handlers, struct {
		Pattern      string
		HandlerFunc  WorkerHandlerFunc
		ErrorHandler []WorkerErrorHandler
	}{
		Pattern: pattern, HandlerFunc: handlerFunc, ErrorHandler: errHandlers,
	})
}
