package interfaces

import "github.com/Shopify/sarama"

// Broker abstraction
type Broker interface {
	Config() *sarama.Config
	Publisher() Publisher
}
