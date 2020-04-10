package app

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/Shopify/sarama"
	"agungdwiprasetyo.com/backend-microservices/internal/factory/constant"
	"agungdwiprasetyo.com/backend-microservices/internal/factory/interfaces"
)

// KafkaConsumer consume data from kafka
func (a *App) KafkaConsumer() {
	if a.kafkaConsumer == nil {
		return
	}

	var consumeTopics []string
	var handlers = make(map[string][]interfaces.SubscriberDelivery)
	for _, m := range a.modules {
		if h := m.SubscriberHandler(constant.Kafka); h != nil {
			for _, topic := range h.GetTopics() {
				handlers[topic] = append(handlers[topic], h)
			}
			consumeTopics = append(consumeTopics, h.GetTopics()...)
		}
	}
	consumer := kafkaConsumer{
		handlers: handlers,
	}

	fmt.Printf("[KAFKA-TOPIC] --> [%s]\n", strings.Join(consumeTopics, "; "))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := a.kafkaConsumer.Consume(ctx, consumeTopics, &consumer); err != nil {
		log.Printf("Error from consumer: %v", err)
	}
}

// kafkaConsumer represents a Sarama consumer group consumer
type kafkaConsumer struct {
	handlers map[string][]interfaces.SubscriberDelivery
}

// Setup is run at the beginning of a new session, before ConsumeClaim
func (c *kafkaConsumer) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited
func (c *kafkaConsumer) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages().
func (c *kafkaConsumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {

	for message := range claim.Messages() {
		log.Printf("Message claimed: value = %s, timestamp = %v, topic = %s", string(message.Value), message.Timestamp, message.Topic)

		for _, handler := range c.handlers[message.Topic] {
			handler.ProcessMessage(session.Context(), message.Value)
		}

		session.MarkMessage(message, "")
	}

	return nil
}
