package rabbitmqworker

import (
	"fmt"

	"github.com/streadway/amqp"
	"pkg.agungdp.dev/candi/config/env"
)

func setupQueueConfig(ch *amqp.Channel, queueName string) (<-chan amqp.Delivery, error) {
	queue, err := ch.QueueDeclare(queueName, true, false, false, false, nil)
	if err != nil {
		return nil, fmt.Errorf("error in declaring the queue %s", err)
	}
	if err := ch.QueueBind(queue.Name, queue.Name, env.BaseEnv().RabbitMQ.ExchangeName, false, nil); err != nil {
		return nil, fmt.Errorf("Queue bind error: %s", err)
	}
	return ch.Consume(
		queue.Name,
		env.BaseEnv().RabbitMQ.ConsumerGroup+"_"+queue.Name, // consumer
		false, // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
}
