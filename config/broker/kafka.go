package broker

import "github.com/Shopify/sarama"

// InitKafkaConfig init kafka broker configuration
func InitKafkaConfig() *sarama.Config {
	kafkaConsumerConfig := sarama.NewConfig()
	kafkaConsumerConfig.Version, _ = sarama.ParseKafkaVersion("2.1.1")
	kafkaConsumerConfig.Consumer.Offsets.Initial = sarama.OffsetOldest

	return kafkaConsumerConfig
}
