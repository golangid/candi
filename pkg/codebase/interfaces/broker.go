package interfaces

import (
	"context"

	"github.com/Shopify/sarama"
)

// Broker abstraction
type Broker interface {
	GetConfig() *sarama.Config
	Publisher() Publisher
	Disconnect(ctx context.Context) error
}
