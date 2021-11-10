package appfactory

import (
	rabbitmqworker "github.com/golangid/candi/codebase/app/rabbitmq_worker"
	"github.com/golangid/candi/codebase/factory"
	"github.com/golangid/candi/config/env"
)

// SetupRabbitMQWorker setup cron worker with default config
func SetupRabbitMQWorker(service factory.ServiceFactory) factory.AppServerFactory {
	return rabbitmqworker.NewWorker(service,
		rabbitmqworker.SetMaxGoroutines(env.BaseEnv().MaxGoroutines),
		rabbitmqworker.SetDebugMode(env.BaseEnv().DebugMode),
		rabbitmqworker.SetConsumerGroup(env.BaseEnv().RabbitMQ.ConsumerGroup),
		rabbitmqworker.SetExchangeName(env.BaseEnv().RabbitMQ.ExchangeName),
		rabbitmqworker.SetBrokerHost(env.BaseEnv().RabbitMQ.Broker),
	)
}
