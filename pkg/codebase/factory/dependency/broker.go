package dependency

import (
	"agungdwiprasetyo.com/backend-microservices/pkg/codebase/interfaces"
	"github.com/Shopify/sarama"
)

type broker struct {
	cfg *sarama.Config
	pub interfaces.Publisher
}

func initBroker(cfg *sarama.Config, pub interfaces.Publisher) interfaces.Broker {
	return &broker{
		cfg: cfg,
		pub: pub,
	}
}

func (b *broker) Config() *sarama.Config {
	return b.cfg
}
func (b *broker) Publisher() interfaces.Publisher {
	return b.pub
}
