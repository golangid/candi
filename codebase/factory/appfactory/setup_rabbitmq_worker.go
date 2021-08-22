package appfactory

import (
	rabbitmqworker "pkg.agungdp.dev/candi/codebase/app/rabbitmq_worker"
	"pkg.agungdp.dev/candi/codebase/factory"
	"pkg.agungdp.dev/candi/config/env"
)

func setupRabbitMQWorker(service factory.ServiceFactory) factory.AppServerFactory {
	return rabbitmqworker.NewWorker(service,
		rabbitmqworker.SetMaxGoroutines(env.BaseEnv().MaxGoroutines),
		rabbitmqworker.SetDebugMode(env.BaseEnv().DebugMode),
		rabbitmqworker.SetConsumerGroup(env.BaseEnv().RabbitMQ.ConsumerGroup),
		rabbitmqworker.SetExchangeName(env.BaseEnv().RabbitMQ.ExchangeName),
		rabbitmqworker.SetBrokerHost(env.BaseEnv().RabbitMQ.Broker),
	)
}
