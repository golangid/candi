package broker

import (
	"context"
	"os"
	"time"

	"github.com/Shopify/sarama"
	"pkg.agungdwiprasetyo.com/candi/codebase/interfaces"
	"pkg.agungdwiprasetyo.com/candi/logger"
	"pkg.agungdwiprasetyo.com/candi/publisher"
)

type kafkaBroker struct {
	client sarama.Client
	pub    interfaces.Publisher
}

func (b *kafkaBroker) GetClient() sarama.Client {
	return b.client
}
func (b *kafkaBroker) Publisher() interfaces.Publisher {
	return b.pub
}
func (b *kafkaBroker) Disconnect(ctx context.Context) error {
	deferFunc := logger.LogWithDefer("kafka: disconnect...")
	defer deferFunc()

	return b.client.Close()
}

// InitKafkaBroker init kafka broker configuration
func InitKafkaBroker(brokers []string, clientID string) interfaces.Broker {
	deferFunc := logger.LogWithDefer("Load Kafka configuration... ")
	defer deferFunc()

	version, ok := os.LookupEnv("KAFKA_CLIENT_VERSION")
	if !ok {
		version = "2.0.0"
	}

	kafkaConfig := sarama.NewConfig()
	kafkaConfig.Version, _ = sarama.ParseKafkaVersion(version)

	// Producer config
	kafkaConfig.ClientID = clientID
	kafkaConfig.Producer.Retry.Max = 15
	kafkaConfig.Producer.Retry.Backoff = 50 * time.Millisecond
	kafkaConfig.Producer.RequiredAcks = sarama.WaitForAll
	kafkaConfig.Producer.Return.Successes = true

	// Consumer config
	kafkaConfig.Consumer.Offsets.Initial = sarama.OffsetOldest

	saramaClient, err := sarama.NewClient(brokers, kafkaConfig)
	if err != nil {
		panic(err)
	}

	return &kafkaBroker{
		client: saramaClient,
		pub:    publisher.NewKafkaPublisher(saramaClient),
	}
}
