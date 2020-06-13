package constant

// Worker is the type returned by a classifier worker (kafka, redis, rabbitmq)
type Worker int

const (
	// Kafka subscriber
	Kafka Worker = iota
	// Redis subscriber
	Redis
	// RabbitMQ subscriber
	RabbitMQ
)
