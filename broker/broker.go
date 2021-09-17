package broker

import (
	"context"

	"github.com/golangid/candi/candihelper"
	"github.com/golangid/candi/codebase/factory/types"
	"github.com/golangid/candi/codebase/interfaces"
)

// Broker model
type Broker struct {
	brokers map[types.Worker]interfaces.Broker
}

/*
InitBrokers register all broker for publisher or consumer

* for Kafka, pass NewKafkaBroker(...KafkaOptionFunc) in param, init kafka broker configuration from env
KAFKA_BROKERS, KAFKA_CLIENT_ID, KAFKA_CLIENT_VERSION

* for RabbitMQ, pass NewRabbitMQBroker(...RabbitMQOptionFunc) in param, init rabbitmq broker configuration from env
RABBITMQ_BROKER, RABBITMQ_CONSUMER_GROUP, RABBITMQ_EXCHANGE_NAME
*/
func InitBrokers(brokers ...interfaces.Broker) *Broker {
	brokerInst := &Broker{
		brokers: make(map[types.Worker]interfaces.Broker),
	}
	for _, bk := range brokers {
		if _, ok := brokerInst.brokers[bk.GetName()]; ok {
			panic("Register broker: " + bk.GetName() + " has been registered")
		}
		brokerInst.brokers[bk.GetName()] = bk
	}

	return brokerInst
}

// GetBrokers get all registered broker
func (b *Broker) GetBrokers() map[types.Worker]interfaces.Broker {
	return b.brokers
}

// RegisterBroker register new broker
func (b *Broker) RegisterBroker(brokerName types.Worker, bk interfaces.Broker) {
	if b.brokers == nil {
		b.brokers = make(map[types.Worker]interfaces.Broker)
	}

	if _, ok := b.brokers[brokerName]; ok {
		panic("Register broker: " + brokerName + " has been registered")
	}
	b.brokers[brokerName] = bk
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
