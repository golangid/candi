package broker

import (
	"context"
	"errors"

	"github.com/Shopify/sarama"
	"pkg.agungdwiprasetyo.com/candi/candihelper"
	"pkg.agungdwiprasetyo.com/candi/codebase/factory/types"
	"pkg.agungdwiprasetyo.com/candi/codebase/interfaces"
	"pkg.agungdwiprasetyo.com/candi/logger"
)

type brokerInstance struct {
	kafka *kafkaBroker
}

// InitBrokers init registered broker
// for kafka pass types.Kafka in param, init kafka broker configuration from env KAFKA_BROKERS, KAFKA_CLIENT_ID, KAFKA_CLIENT_VERSION
func InitBrokers(brokerTypes ...types.Worker) interfaces.Broker {
	var brokerInst = brokerInstance{
		kafka: &kafkaBroker{},
	}
	for _, brokerType := range brokerTypes {
		switch brokerType {
		case types.Kafka:
			brokerInst.kafka = initKafkaBroker()
		}
	}
	return &brokerInst
}

func (b *brokerInstance) GetKafkaClient() sarama.Client {
	return b.kafka.client
}

func (b *brokerInstance) Publisher(brokerType types.Worker) interfaces.Publisher {
	switch brokerType {
	case types.Kafka:
		return b.kafka.pub
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

	if mErr.HasError() {
		return mErr
	}
	return nil
}
