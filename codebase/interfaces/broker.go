package interfaces

import (
	"github.com/Shopify/sarama"
	"github.com/streadway/amqp"
	"pkg.agungdp.dev/candi/codebase/factory/types"
)

// Broker abstraction
type Broker interface {
	GetKafkaClient() sarama.Client
	GetRabbitMQConn() *amqp.Connection
	Publisher(types.Worker) Publisher
	Health() map[string]error
	Closer
}
