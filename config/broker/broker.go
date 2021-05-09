package broker

import (
	"context"
	"errors"

	"pkg.agungdp.dev/candi/candihelper"
	"pkg.agungdp.dev/candi/codebase/factory/types"
	"pkg.agungdp.dev/candi/codebase/interfaces"
	"pkg.agungdp.dev/candi/logger"
)

// OptionFunc type
type OptionFunc func(*brokerInstance)

// SetKafka set kafka broker
func SetKafka(bk *KafkaBroker) OptionFunc {
	return func(bi *brokerInstance) {
		bi.kafka = bk
	}
}

// SetRabbitMQ set kafka broker
func SetRabbitMQ(bk *RabbitMQBroker) OptionFunc {
	return func(bi *brokerInstance) {
		bi.rabbitmq = bk
	}
}

type brokerInstance struct {
	kafka    *KafkaBroker
	rabbitmq *RabbitMQBroker
}

/*
InitBrokers init registered broker

* for kafka, pass NewKafkaBroker(...opts) in param, init kafka broker configuration from env
KAFKA_BROKERS, KAFKA_CLIENT_ID, KAFKA_CLIENT_VERSION

* for rabbitmq, pass NewRabbitMQBroker() in param, init rabbitmq broker configuration from env
RABBITMQ_BROKER, RABBITMQ_CONSUMER_GROUP, RABBITMQ_EXCHANGE_NAME
*/
func InitBrokers(opts ...OptionFunc) interfaces.Broker {
	brokerInst := new(brokerInstance)
	for _, opt := range opts {
		opt(brokerInst)
	}

	return brokerInst
}

func (b *brokerInstance) GetConfiguration(brokerType types.Worker) interface{} {
	switch brokerType {
	case types.Kafka:
		return b.kafka.client
	case types.RabbitMQ:
		return b.rabbitmq.conn
	}
	return nil
}

func (b *brokerInstance) Publisher(brokerType types.Worker) interfaces.Publisher {
	switch brokerType {
	case types.Kafka:
		return b.kafka.pub
	case types.RabbitMQ:
		return b.rabbitmq.pub
	}
	return nil
}

func (b *brokerInstance) Health() map[string]error {
	mErr := make(map[string]error)

	if b.kafka.client != nil {
		var err error
		if len(b.kafka.client.Brokers()) == 0 {
			err = errors.New("not ok")
		}
		mErr[string(types.Kafka)] = err
	}

	if b.rabbitmq.conn != nil {
		var err error
		mErr[string(types.RabbitMQ)] = err
	}

	return mErr
}

func (b *brokerInstance) Disconnect(ctx context.Context) error {

	mErr := candihelper.NewMultiError()
	if b.kafka != nil {
		func() {
			deferFunc := logger.LogWithDefer("kafka: disconnect...")
			defer deferFunc()
			if err := b.kafka.client.Close(); err != nil {
				mErr.Append(string(types.Kafka), err)
			}
		}()
	}

	if b.rabbitmq != nil {
		func() {
			deferFunc := logger.LogWithDefer("rabbitmq: disconnect...")
			defer deferFunc()
			if err := b.rabbitmq.conn.Close(); err != nil {
				mErr.Append(string(types.RabbitMQ), err)
			}
		}()
	}

	if mErr.HasError() {
		return mErr
	}
	return nil
}
