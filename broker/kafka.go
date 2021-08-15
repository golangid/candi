package broker

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Shopify/sarama"
	"pkg.agungdp.dev/candi/candihelper"
	"pkg.agungdp.dev/candi/candishared"
	"pkg.agungdp.dev/candi/codebase/factory/types"
	"pkg.agungdp.dev/candi/codebase/interfaces"
	"pkg.agungdp.dev/candi/config/env"
	"pkg.agungdp.dev/candi/logger"
	"pkg.agungdp.dev/candi/tracer"
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
		kb.publisher = pub
	}
}

// KafkaBroker configuration
type KafkaBroker struct {
	config    *sarama.Config
	client    sarama.Client
	publisher interfaces.Publisher
}

// NewKafkaBroker setup kafka configuration for publisher or consumer, empty option param for default configuration
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

	if kb.publisher == nil {
		kb.publisher = NewKafkaPublisher(saramaClient, false) // default publisher is sync
	}

	return kb
}

// GetConfiguration method
func (k *KafkaBroker) GetConfiguration() interface{} {
	return k.client
}

// GetPublisher method
func (k *KafkaBroker) GetPublisher() interfaces.Publisher {
	return k.publisher
}

// Health method
func (k *KafkaBroker) Health() map[string]error {
	mErr := make(map[string]error)

	var err error
	if len(k.client.Brokers()) == 0 {
		err = errors.New("not ok")
	}
	mErr[string(types.Kafka)] = err

	return mErr
}

// Disconnect method
func (k *KafkaBroker) Disconnect(ctx context.Context) error {
	deferFunc := logger.LogWithDefer("kafka: disconnect...")
	defer deferFunc()

	return k.client.Close()
}

// kafkaPublisher kafka publisher
type kafkaPublisher struct {
	producerSync  sarama.SyncProducer
	producerAsync sarama.AsyncProducer
}

// NewKafkaPublisher setup only kafka publisher with client connection
func NewKafkaPublisher(client sarama.Client, async bool) interfaces.Publisher {
	var err error

	kafkaPublisher := &kafkaPublisher{}
	if async {
		kafkaPublisher.producerAsync, err = sarama.NewAsyncProducerFromClient(client)
	} else {
		kafkaPublisher.producerSync, err = sarama.NewSyncProducerFromClient(client)
	}

	if err != nil {
		logger.LogYellow(fmt.Sprintf("(Kafka publisher: warning, %v. Should be panicked when using kafka publisher.) ", err))
		return nil
	}

	return kafkaPublisher
}

// PublishMessage method
func (p *kafkaPublisher) PublishMessage(ctx context.Context, args *candishared.PublisherArgument) (err error) {
	trace := tracer.StartTrace(ctx, "kafka:publish_message")
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
		trace.SetError(err)
		trace.Finish()
	}()

	payload := candihelper.ToBytes(args.Data)

	trace.SetTag("topic", args.Topic)
	trace.SetTag("key", args.Key)
	trace.Log("message", payload)

	msg := &sarama.ProducerMessage{
		Topic:     args.Topic,
		Key:       sarama.ByteEncoder([]byte(args.Key)),
		Value:     sarama.ByteEncoder(payload),
		Timestamp: time.Now(),
	}

	if p.producerSync != nil {
		_, _, err = p.producerSync.SendMessage(msg)
	} else {
		p.producerAsync.Input() <- msg
	}
	return
}
