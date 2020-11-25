package interfaces

import (
	"github.com/Shopify/sarama"
	"pkg.agungdwiprasetyo.com/candi/codebase/factory/types"
)

// Broker abstraction
type Broker interface {
	GetKafkaClient() sarama.Client
	Publisher(types.Worker) Publisher
	Health() map[string]error
	Closer
}
