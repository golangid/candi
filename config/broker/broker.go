package broker

import (
	"context"
	"errors"

	"github.com/Shopify/sarama"
	"github.com/streadway/amqp"
	"pkg.agungdp.dev/candi/candihelper"
	"pkg.agungdp.dev/candi/codebase/factory/types"
	"pkg.agungdp.dev/candi/codebase/interfaces"
	"pkg.agungdp.dev/candi/logger"
)

type brokerInstance struct {
	kafka    *kafkaBroker
	rabbitmq *rabbitmqBroker
}

/*
InitBrokers init registered broker

* for kafka, pass types.Kafka in param, init kafka broker configuration from env
KAFKA_BROKERS, KAFKA_CLIENT_ID, KAFKA_CLIENT_VERSION

* for rabbitmq, pass types.RabbitMQ in param, init rabbitmq broker configuration from env
RABBITMQ_BROKER, RABBITMQ_CONSUMER_GROUP, RABBITMQ_EXCHANGE_NAME
*/
func InitBrokers(brokerTypes ...types.Worker) interfaces.Broker {
	var brokerInst = brokerInstance{
		kafka:    &kafkaBroker{},
		rabbitmq: &rabbitmqBroker{},
	}
	for _, brokerType := range brokerTypes {
		switch brokerType {
		case types.Kafka:
			brokerInst.kafka = initKafkaBroker()
		case types.RabbitMQ:
			brokerInst.rabbitmq = initRabbitMQBroker()
		}
	}
	return &brokerInst
}

func (b *brokerInstance) GetKafkaClient() sarama.Client {
	return b.kafka.client
}

func (b *brokerInstance) GetRabbitMQConn() *amqp.Connection {
	return b.rabbitmq.conn
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
	if b.kafka.client != nil {
		func() {
			deferFunc := logger.LogWithDefer("kafka: disconnect...")
			defer deferFunc()
			if err := b.kafka.client.Close(); err != nil {
				mErr.Append(string(types.Kafka), err)
			}
		}()
	}

	if b.rabbitmq.conn != nil {
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
