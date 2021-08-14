package broker

import (
	"context"

	"pkg.agungdp.dev/candi/candihelper"
	"pkg.agungdp.dev/candi/codebase/factory/types"
	"pkg.agungdp.dev/candi/codebase/interfaces"
)

// OptionFunc type
type OptionFunc func(*Broker)

// SetKafka setup kafka broker for publisher or consumer
func SetKafka(bk interfaces.Broker) OptionFunc {
	return func(bi *Broker) {
		bi.brokers[types.Kafka] = bk
	}
}

// SetRabbitMQ setup rabbitmq broker for publisher or consumer
func SetRabbitMQ(bk interfaces.Broker) OptionFunc {
	return func(bi *Broker) {
		bi.brokers[types.RabbitMQ] = bk
	}
}

// Broker model
type Broker struct {
	brokers map[types.Worker]interfaces.Broker
}

/*
InitBrokers init registered broker for publisher or consumer

* for kafka, pass NewKafkaBroker(...KafkaOptionFunc) in param, init kafka broker configuration from env
KAFKA_BROKERS, KAFKA_CLIENT_ID, KAFKA_CLIENT_VERSION

* for rabbitmq, pass NewRabbitMQBroker(...RabbitMQOptionFunc) in param, init rabbitmq broker configuration from env
RABBITMQ_BROKER, RABBITMQ_CONSUMER_GROUP, RABBITMQ_EXCHANGE_NAME
*/
func InitBrokers(opts ...OptionFunc) *Broker {
	brokerInst := &Broker{
		brokers: make(map[types.Worker]interfaces.Broker),
	}
	for _, opt := range opts {
		opt(brokerInst)
	}

	return brokerInst
}

// GetBrokers get all registered broker
func (b *Broker) GetBrokers() map[types.Worker]interfaces.Broker {
	return b.brokers
}

// Disconnect disconnect all registered broker
func (b *Broker) Disconnect(ctx context.Context) error {
	mErr := candihelper.NewMultiError()

	for name, broker := range b.brokers {
		mErr.Append(string(name), broker.Disconnect(ctx))
	}

	if mErr.HasError() {
		return mErr
	}
	return nil
}
