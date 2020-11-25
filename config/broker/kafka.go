package broker

import (
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

func initKafkaBroker() *kafkaBroker {
	deferFunc := logger.LogWithDefer("Load Kafka broker configuration... ")
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
