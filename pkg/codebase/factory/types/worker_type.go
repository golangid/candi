package types

import "context"

// Worker is the type returned by a classifier worker (kafka, redis subscriber, rabbitmq, scheduler, task queue)
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

// WorkerHandlerGroup group of worker handlers by pattern string
type WorkerHandlerGroup struct {
	Handlers []struct {
		Pattern     string
		HandlerFunc WorkerHandlerFunc
	}
}

// Add method from WorkerHandlerGroup, pattern can contains unique topic name, key, and task name
func (m *WorkerHandlerGroup) Add(pattern string, handlerFunc WorkerHandlerFunc) {
	m.Handlers = append(m.Handlers, struct {
		Pattern     string
		HandlerFunc WorkerHandlerFunc
	}{
		Pattern: pattern, HandlerFunc: handlerFunc,
	})
}
