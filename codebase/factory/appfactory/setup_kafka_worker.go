package appfactory

import (
	kafkaworker "github.com/golangid/candi/codebase/app/kafka_worker"
	"github.com/golangid/candi/codebase/factory"
	"github.com/golangid/candi/config/env"
)

func setupKafkaWorker(service factory.ServiceFactory) factory.AppServerFactory {
	return kafkaworker.NewWorker(service,
		kafkaworker.SetMaxGoroutines(env.BaseEnv().MaxGoroutines),
		kafkaworker.SetDebugMode(env.BaseEnv().DebugMode),
		kafkaworker.SetConsumerGroup(env.BaseEnv().Kafka.ConsumerGroup),
		kafkaworker.SetBrokers(env.BaseEnv().Kafka.Brokers),
	)
}
