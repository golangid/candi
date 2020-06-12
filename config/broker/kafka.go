package broker

import (
	"time"

	"github.com/Shopify/sarama"
)

// InitKafkaConfig init kafka broker configuration
func InitKafkaConfig(isUseConsumer bool, clientID string) *sarama.Config {

	kafkaConfig := sarama.NewConfig()
	kafkaConfig.Version, _ = sarama.ParseKafkaVersion("2.1.1")

	// Producer config
	kafkaConfig.ClientID = clientID
	kafkaConfig.Producer.Retry.Max = 15
	kafkaConfig.Producer.Retry.Backoff = 50 * time.Millisecond
	kafkaConfig.Producer.RequiredAcks = sarama.WaitForAll
	kafkaConfig.Producer.Return.Successes = true

	if !isUseConsumer {
		// Consumer config
		kafkaConfig.Consumer.Offsets.Initial = sarama.OffsetOldest
	}

	return kafkaConfig
}
