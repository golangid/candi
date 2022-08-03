package broker

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Shopify/sarama"
	"github.com/golangid/candi/candihelper"
	"github.com/golangid/candi/candishared"
	"github.com/golangid/candi/codebase/factory/types"
	"github.com/golangid/candi/codebase/interfaces"
	"github.com/golangid/candi/config/env"
	"github.com/golangid/candi/logger"
	"github.com/golangid/candi/tracer"
)

// KafkaOptionFunc func type
type KafkaOptionFunc func(*KafkaBroker)

// KafkaSetBrokerHost set custom broker host
func KafkaSetBrokerHost(brokers []string) KafkaOptionFunc {
	return func(kb *KafkaBroker) {
		kb.brokerHost = brokers
	}
}

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

// GetDefaultKafkaConfig construct default kafka config
func GetDefaultKafkaConfig(additionalConfigFunc ...func(*sarama.Config)) *sarama.Config {
	version := env.BaseEnv().Kafka.ClientVersion
	if version == "" {
		version = "2.0.0"
	}

	// set default configuration
	cfg := sarama.NewConfig()
	cfg.Version, _ = sarama.ParseKafkaVersion(version)

	// Producer config
	cfg.ClientID = env.BaseEnv().Kafka.ClientID
	cfg.Producer.Retry.Max = 15
	cfg.Producer.Retry.Backoff = 50 * time.Millisecond
	cfg.Producer.RequiredAcks = sarama.WaitForAll
	cfg.Producer.Return.Successes = true

	// Consumer config
	cfg.Consumer.Offsets.Initial = sarama.OffsetOldest
	cfg.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin

	for _, additionalFunc := range additionalConfigFunc {
		additionalFunc(cfg)
	}

	return cfg
}

// KafkaBroker configuration
type KafkaBroker struct {
	brokerHost []string
	config     *sarama.Config
	client     sarama.Client
	publisher  interfaces.Publisher
}

// NewKafkaBroker setup kafka configuration for publisher or consumer, empty option param for default configuration
func NewKafkaBroker(opts ...KafkaOptionFunc) *KafkaBroker {
	deferFunc := logger.LogWithDefer("Load Kafka broker configuration... ")
	defer deferFunc()

	kb := new(KafkaBroker)
	kb.brokerHost = env.BaseEnv().Kafka.Brokers
	for _, opt := range opts {
		opt(kb)
	}

	if kb.config == nil {
		// set default configuration
		kb.config = GetDefaultKafkaConfig()
	}

	saramaClient, err := sarama.NewClient(kb.brokerHost, kb.config)
	if err != nil {
		panic(fmt.Errorf("%s. Brokers: %s", err, strings.Join(kb.brokerHost, ", ")))
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

// GetName method
func (k *KafkaBroker) GetName() types.Worker {
	return types.Kafka
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

	var payload []byte
	if len(args.Message) > 0 {
		payload = args.Message
	} else {
		payload = candihelper.ToBytes(args.Data)
	}

	trace.SetTag("topic", args.Topic)
	trace.SetTag("key", args.Key)
	trace.Log("header", args.Header)
	trace.Log("message", payload)

	msg := &sarama.ProducerMessage{
		Topic:     args.Topic,
		Key:       sarama.ByteEncoder([]byte(args.Key)),
		Value:     sarama.ByteEncoder(payload),
		Timestamp: time.Now(),
	}

	traceHeader := map[string]string{}
	trace.InjectRequestHeader(traceHeader)
	for k, v := range traceHeader {
		msg.Headers = append(msg.Headers, sarama.RecordHeader{
			Key:   []byte(k),
			Value: []byte(v),
		})
	}

	for keyHeader, valueHeader := range args.Header {
		msg.Headers = append(msg.Headers, sarama.RecordHeader{
			Key:   []byte(keyHeader),
			Value: candihelper.ToBytes(valueHeader),
		})
	}

	if p.producerSync != nil {
		_, _, err = p.producerSync.SendMessage(msg)
	} else {
		p.producerAsync.Input() <- msg
	}
	return
}
