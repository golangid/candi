package broker

import (
	"time"

	"github.com/Shopify/sarama"
	"pkg.agungdp.dev/candi/codebase/interfaces"
	"pkg.agungdp.dev/candi/config/env"
	"pkg.agungdp.dev/candi/logger"
	"pkg.agungdp.dev/candi/publisher"
)

// KafkaOptionFunc func type
type KafkaOptionFunc func(*KafkaBroker)

// KafkaSetConfig set custom sarama configuration
func KafkaSetConfig(cfg *sarama.Config) KafkaOptionFunc {
	return func(kb *KafkaBroker) {
		kb.config = cfg
	}
}

// KafkaSetPublisher set custom publisher
func KafkaSetPublisher(pub interfaces.Publisher) KafkaOptionFunc {
	return func(kb *KafkaBroker) {
		kb.pub = pub
	}
}

// KafkaBroker configuration
type KafkaBroker struct {
	config *sarama.Config
	client sarama.Client
	pub    interfaces.Publisher
}

// NewKafkaBroker constructor with option, empty option param for default configuration
func NewKafkaBroker(opts ...KafkaOptionFunc) *KafkaBroker {
	deferFunc := logger.LogWithDefer("Load Kafka broker configuration... ")
	defer deferFunc()

	kb := new(KafkaBroker)
	for _, opt := range opts {
		opt(kb)
	}

	version := env.BaseEnv().Kafka.ClientVersion
	if version == "" {
		version = "2.0.0"
	}

	if kb.config == nil {
		// set default configuration
		kb.config = sarama.NewConfig()
		kb.config.Version, _ = sarama.ParseKafkaVersion(version)

		// Producer config
		kb.config.ClientID = env.BaseEnv().Kafka.ClientID
		kb.config.Producer.Retry.Max = 15
		kb.config.Producer.Retry.Backoff = 50 * time.Millisecond
		kb.config.Producer.RequiredAcks = sarama.WaitForAll
		kb.config.Producer.Return.Successes = true

		// Consumer config
		kb.config.Consumer.Offsets.Initial = sarama.OffsetOldest
		kb.config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	}

	saramaClient, err := sarama.NewClient(env.BaseEnv().Kafka.Brokers, kb.config)
	if err != nil {
		panic(err)
	}
	kb.client = saramaClient

	if kb.pub == nil {
		kb.pub = publisher.NewKafkaPublisher(saramaClient)
	}

	return kb
}
