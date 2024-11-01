package broker

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/IBM/sarama"
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

// KafkaSetWorkerType set worker type
func KafkaSetWorkerType(workerType types.Worker) KafkaOptionFunc {
	return func(bk *KafkaBroker) {
		bk.WorkerType = workerType
	}
}

// KafkaSetBrokerHost set custom broker host
func KafkaSetBrokerHost(brokers []string) KafkaOptionFunc {
	return func(kb *KafkaBroker) {
		kb.BrokerHost = brokers
	}
}

// KafkaSetConfig set custom sarama configuration
func KafkaSetConfig(cfg *sarama.Config) KafkaOptionFunc {
	return func(kb *KafkaBroker) {
		kb.Config = cfg
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
	cfg.ClientID = env.BaseEnv().Kafka.ClientID

	// Producer config
	cfg.Producer.Retry.Max = 15
	cfg.Producer.Retry.Backoff = 50 * time.Millisecond
	cfg.Producer.RequiredAcks = sarama.WaitForAll
	cfg.Producer.Return.Successes = true

	// Consumer config
	cfg.Consumer.Offsets.Initial = sarama.OffsetOldest
	cfg.Consumer.Group.Rebalance.Strategy = sarama.NewBalanceStrategyRoundRobin()

	for _, additionalFunc := range additionalConfigFunc {
		additionalFunc(cfg)
	}

	return cfg
}

// KafkaBroker configuration
type KafkaBroker struct {
	WorkerType types.Worker
	BrokerHost []string
	Config     *sarama.Config
	Client     sarama.Client
	publisher  interfaces.Publisher
}

// NewKafkaBroker setup kafka configuration for publisher or consumer, empty option param for default configuration (with default worker type is types.Kafka)
func NewKafkaBroker(opts ...KafkaOptionFunc) *KafkaBroker {
	defer logger.LogWithDefer("Load Kafka broker configuration... ")()

	kb := new(KafkaBroker)
	kb.BrokerHost = env.BaseEnv().Kafka.Brokers
	kb.WorkerType = types.Kafka
	for _, opt := range opts {
		opt(kb)
	}

	if kb.Config == nil {
		// set default configuration
		kb.Config = GetDefaultKafkaConfig()
	}

	saramaClient, err := sarama.NewClient(kb.BrokerHost, kb.Config)
	if err != nil {
		panic(fmt.Errorf("%s. Brokers: %s", err, strings.Join(kb.BrokerHost, ", ")))
	}
	kb.Client = saramaClient

	if kb.publisher == nil {
		kb.publisher = NewKafkaPublisher(saramaClient, false) // default publisher is sync
	}

	return kb
}

// GetPublisher method
func (k *KafkaBroker) GetPublisher() interfaces.Publisher {
	return k.publisher
}

// GetName method
func (k *KafkaBroker) GetName() types.Worker {
	return k.WorkerType
}

// Health method
func (k *KafkaBroker) Health() map[string]error {
	mErr := make(map[string]error)

	var err error
	if len(k.Client.Brokers()) == 0 {
		err = errors.New("not ok")
	}
	mErr[string(types.Kafka)] = err

	return mErr
}

// Disconnect method
func (k *KafkaBroker) Disconnect(ctx context.Context) error {
	defer logger.LogWithDefer("\x1b[33;5mkafka_broker\x1b[0m: disconnect...")()

	return k.Client.Close()
}

// kafkaPublisher kafka publisher
type kafkaPublisher struct {
	producerSync  sarama.SyncProducer
	producerAsync sarama.AsyncProducer
	broker        string
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

	brokers := client.Brokers()
	for i, cl := range brokers {
		kafkaPublisher.broker += cl.Addr()
		if i < len(brokers)-1 {
			kafkaPublisher.broker += ","
		}
	}
	return kafkaPublisher
}

// PublishMessage method
func (p *kafkaPublisher) PublishMessage(ctx context.Context, args *candishared.PublisherArgument) (err error) {
	trace, _ := tracer.StartTraceWithContext(ctx, "kafka:publish_message")
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
		trace.Finish(tracer.FinishWithError(err))
	}()

	var payload []byte
	if len(args.Message) > 0 {
		payload = args.Message
	} else {
		payload = candihelper.ToBytes(args.Data)
	}

	trace.SetTag("brokers", p.broker)
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
	if !args.Timestamp.IsZero() {
		msg.Timestamp = args.Timestamp
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
		trace.SetError(err)
	} else {
		p.producerAsync.Input() <- msg
	}
	return
}
