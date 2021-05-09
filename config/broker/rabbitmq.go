package broker

import (
	"github.com/streadway/amqp"
	"pkg.agungdp.dev/candi/codebase/interfaces"
	"pkg.agungdp.dev/candi/config/env"
	"pkg.agungdp.dev/candi/logger"
	"pkg.agungdp.dev/candi/publisher"
)

// RabbitMQBroker broker
type RabbitMQBroker struct {
	conn *amqp.Connection
	pub  interfaces.Publisher
}

// NewRabbitMQBroker constructor, connection from RABBITMQ_BROKER environment
func NewRabbitMQBroker() *RabbitMQBroker {
	deferFunc := logger.LogWithDefer("Load RabbitMQ broker configuration... ")
	defer deferFunc()

	conn, err := amqp.Dial(env.BaseEnv().RabbitMQ.Broker)
	if err != nil {
		panic("RabbitMQ: cannot connect to server broker: " + err.Error())
	}
	return &RabbitMQBroker{
		conn: conn,
		pub:  publisher.NewRabbitMQPublisher(conn),
	}
}
