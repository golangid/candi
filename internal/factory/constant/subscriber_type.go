package constant

// Subscriber is the type returned by a classifier subscriber (kafka, redis, rabbitmq)
type Subscriber int

const (
	// Kafka subscriber
	Kafka Subscriber = iota
	// Redis subscriber
	Redis
	// RabbitMQ subscriber
	RabbitMQ
)
