package constant

// Worker is the type returned by a classifier worker (kafka, redis, rabbitmq, scheduler)
type Worker int

const (
	// Kafka worker
	Kafka Worker = iota
	// Redis worker
	Redis
	// RabbitMQ worker
	RabbitMQ
	// Scheduler worker
	Scheduler
)
