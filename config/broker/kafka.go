package broker

import (
	"context"
	"time"

	"agungdwiprasetyo.com/backend-microservices/pkg/codebase/interfaces"
	"agungdwiprasetyo.com/backend-microservices/pkg/logger"
	"agungdwiprasetyo.com/backend-microservices/pkg/publisher"
	"github.com/Shopify/sarama"
)

type kafkaBroker struct {
	cfg *sarama.Config
	pub interfaces.Publisher
}

func (b *kafkaBroker) GetConfig() *sarama.Config {
	return b.cfg
}
func (b *kafkaBroker) Publisher() interfaces.Publisher {
	return b.pub
}
func (b *kafkaBroker) Disconnect(ctx context.Context) error {
	return nil
}

// InitKafkaBroker init kafka broker configuration
func InitKafkaBroker(clientID string) interfaces.Broker {
	deferFunc := logger.LogWithDefer("Load Kafka configuration...")
	defer deferFunc()

	kafkaConfig := sarama.NewConfig()
	kafkaConfig.Version, _ = sarama.ParseKafkaVersion("2.1.1")

	// Producer config
	kafkaConfig.ClientID = clientID
	kafkaConfig.Producer.Retry.Max = 15
	kafkaConfig.Producer.Retry.Backoff = 50 * time.Millisecond
	kafkaConfig.Producer.RequiredAcks = sarama.WaitForAll
	kafkaConfig.Producer.Return.Successes = true

	// Consumer config
	kafkaConfig.Consumer.Offsets.Initial = sarama.OffsetOldest

	return &kafkaBroker{
		cfg: kafkaConfig,
		pub: publisher.NewKafkaPublisher(kafkaConfig),
	}
}
