package broker

import (
	"context"
	"errors"
	"time"

	"github.com/Shopify/sarama"
	"pkg.agungdwiprasetyo.com/candi/codebase/interfaces"
	"pkg.agungdwiprasetyo.com/candi/config/env"
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
func (b *kafkaBroker) Health() map[string]error {
	mErr := make(map[string]error)
	var err error
	if len(b.client.Brokers()) == 0 {
		err = errors.New("not ok")
	}
	mErr["kafka"] = err
	return mErr
}
func (b *kafkaBroker) Disconnect(ctx context.Context) error {
	deferFunc := logger.LogWithDefer("kafka: disconnect...")
	defer deferFunc()

	return b.client.Close()
}

// InitKafkaBroker init kafka broker configuration from env KAFKA_BROKERS, KAFKA_CLIENT_ID, KAFKA_CLIENT_VERSION
func InitKafkaBroker() interfaces.Broker {
	deferFunc := logger.LogWithDefer("Load Kafka configuration... ")
	defer deferFunc()

	version := env.BaseEnv().Kafka.ClientVersion
	if version == "" {
		version = "2.0.0"
	}

	kafkaConfig := sarama.NewConfig()
	kafkaConfig.Version, _ = sarama.ParseKafkaVersion(version)

	// Producer config
	kafkaConfig.ClientID = env.BaseEnv().Kafka.ClientID
	kafkaConfig.Producer.Retry.Max = 15
	kafkaConfig.Producer.Retry.Backoff = 50 * time.Millisecond
	kafkaConfig.Producer.RequiredAcks = sarama.WaitForAll
	kafkaConfig.Producer.Return.Successes = true

	// Consumer config
	kafkaConfig.Consumer.Offsets.Initial = sarama.OffsetOldest
	kafkaConfig.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin

	saramaClient, err := sarama.NewClient(env.BaseEnv().Kafka.Brokers, kafkaConfig)
	if err != nil {
		panic(err)
	}

	return &kafkaBroker{
		client: saramaClient,
		pub:    publisher.NewKafkaPublisher(saramaClient),
	}
}
