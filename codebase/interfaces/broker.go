package interfaces

import (
	"github.com/Shopify/sarama"
	"pkg.agungdp.dev/candi/codebase/factory/types"
)

// Broker abstraction
type Broker interface {
	GetKafkaClient() sarama.Client
	Publisher(types.Worker) Publisher
	Health() map[string]error
	Closer
}
