package rabbitmqworker

import (
	"fmt"

	"github.com/streadway/amqp"
	"pkg.agungdp.dev/candi/config/env"
)

func getConfiguration(conn *amqp.Connection) (*amqp.Channel, func(channel *amqp.Channel, queueName string) (<-chan amqp.Delivery, error)) {
	ch, err := conn.Channel()
	if err != nil {
		panic("RabbitMQ channel: " + err.Error())
	}
	if err := ch.ExchangeDeclare("amq.direct", "direct", true, false, false, false, nil); err != nil {
		panic("RabbitMQ exchange declare direct: " + err.Error())
	}
	if err := ch.ExchangeDeclare(
		env.BaseEnv().RabbitMQ.ExchangeName, // name
		"x-delayed-message",                 // type
		true,                                // durable
		false,                               // auto-deleted
		false,                               // internal
		false,                               // no-wait
		amqp.Table{
			"x-delayed-type": "direct",
		},
	); err != nil {
		panic("RabbitMQ exchange declare delayed: " + err.Error())
	}
	if err := ch.Qos(2, 0, false); err != nil {
		panic("RabbitMQ Qos: " + err.Error())
	}

	return ch, func(channel *amqp.Channel, queueName string) (<-chan amqp.Delivery, error) {
		queue, err := channel.QueueDeclare(queueName, true, false, false, false, nil)
		if err != nil {
			return nil, fmt.Errorf("error in declaring the queue %s", err)
		}
		if err := channel.QueueBind(queue.Name, queue.Name, env.BaseEnv().RabbitMQ.ExchangeName, false, nil); err != nil {
			return nil, fmt.Errorf("Queue bind error: %s", err)
		}
		return channel.Consume(
			queue.Name,
			env.BaseEnv().RabbitMQ.ConsumerGroup+"_"+queue.Name, // consumer
			true,  // auto-ack
			false, // exclusive
			false, // no-local
			false, // no-wait
			nil,   // args
		)
	}
}
