package rabbitmqworker

import (
	"fmt"

	"github.com/streadway/amqp"
)

func setupQueueConfig(ch *amqp.Channel, consumerGroup, exchangeName, queueName string) (<-chan amqp.Delivery, error) {
	queue, err := ch.QueueDeclare(queueName, true, false, false, false, nil)
	if err != nil {
		return nil, fmt.Errorf("error in declaring the queue %s", err)
	}
	if err := ch.QueueBind(queue.Name, queue.Name, exchangeName, false, nil); err != nil {
		return nil, fmt.Errorf("Queue bind error: %s", err)
	}
	return ch.Consume(
		queue.Name,
		consumerGroup+"_"+queue.Name, // consumer
		false,                        // auto-ack
		false,                        // exclusive
		false,                        // no-local
		false,                        // no-wait
		nil,                          // args
	)
}
