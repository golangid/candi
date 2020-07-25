package interfaces

import (
	"github.com/Shopify/sarama"
)

// Broker abstraction
type Broker interface {
	GetClient() sarama.Client
	Publisher() Publisher
	Closer
}
