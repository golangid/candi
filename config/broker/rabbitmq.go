package broker

import (
	"github.com/streadway/amqp"
	"pkg.agungdp.dev/candi/codebase/interfaces"
	"pkg.agungdp.dev/candi/config/env"
	"pkg.agungdp.dev/candi/logger"
	"pkg.agungdp.dev/candi/publisher"
)

type rabbitmqBroker struct {
	conn *amqp.Connection
	pub  interfaces.Publisher
}

func initRabbitMQBroker() *rabbitmqBroker {
	deferFunc := logger.LogWithDefer("Load RabbitMQ broker configuration... ")
	defer deferFunc()

	conn, err := amqp.Dial(env.BaseEnv().RabbitMQ.Broker)
	if err != nil {
		panic("RabbitMQ: cannot connect to server broker: " + err.Error())
	}
	return &rabbitmqBroker{
		conn: conn,
		pub:  publisher.NewRabbitMQPublisher(conn),
	}
}
