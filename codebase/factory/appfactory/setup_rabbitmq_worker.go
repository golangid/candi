package appfactory

import (
	rabbitmqworker "github.com/golangid/candi/codebase/app/rabbitmq_worker"
	"github.com/golangid/candi/codebase/factory"
	"github.com/golangid/candi/codebase/factory/types"
	"github.com/golangid/candi/config/env"
)

// SetupRabbitMQWorker setup rabbitmq worker with default config
func SetupRabbitMQWorker(service factory.ServiceFactory, opts ...rabbitmqworker.OptionFunc) factory.AppServerFactory {
	rabbitMQOpts := []rabbitmqworker.OptionFunc{
		rabbitmqworker.SetMaxGoroutines(env.BaseEnv().MaxGoroutines),
		rabbitmqworker.SetDebugMode(env.BaseEnv().DebugMode),
		rabbitmqworker.SetConsumerGroup(env.BaseEnv().RabbitMQ.ConsumerGroup),
	}
	rabbitMQOpts = append(rabbitMQOpts, opts...)
	return rabbitmqworker.NewWorker(service, service.GetDependency().GetBroker(types.RabbitMQ), rabbitMQOpts...)
}
