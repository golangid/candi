package appfactory

import (
	kafkaworker "pkg.agungdp.dev/candi/codebase/app/kafka_worker"
	"pkg.agungdp.dev/candi/codebase/factory"
	"pkg.agungdp.dev/candi/config/env"
)

func setupKafkaWorker(service factory.ServiceFactory) factory.AppServerFactory {
	return kafkaworker.NewWorker(service,
		kafkaworker.SetMaxGoroutines(env.BaseEnv().MaxGoroutines),
		kafkaworker.SetDebugMode(env.BaseEnv().DebugMode),
		kafkaworker.SetConsumerGroup(env.BaseEnv().Kafka.ConsumerGroup),
		kafkaworker.SetBrokers(env.BaseEnv().Kafka.Brokers),
	)
}
