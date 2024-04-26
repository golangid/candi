package appfactory

import (
	kafkaworker "github.com/golangid/candi/codebase/app/kafka_worker"
	"github.com/golangid/candi/codebase/factory"
	"github.com/golangid/candi/codebase/factory/types"
	"github.com/golangid/candi/config/env"
)

// SetupKafkaWorker setup kafka worker with default config
func SetupKafkaWorker(service factory.ServiceFactory, opts ...kafkaworker.OptionFunc) factory.AppServerFactory {
	kafkaOpts := []kafkaworker.OptionFunc{
		kafkaworker.SetMaxGoroutines(env.BaseEnv().MaxGoroutines),
		kafkaworker.SetDebugMode(env.BaseEnv().DebugMode),
		kafkaworker.SetConsumerGroup(env.BaseEnv().Kafka.ConsumerGroup),
	}
	kafkaOpts = append(kafkaOpts, opts...)
	return kafkaworker.NewWorker(service, service.GetDependency().GetBroker(types.Kafka), kafkaOpts...)
}
