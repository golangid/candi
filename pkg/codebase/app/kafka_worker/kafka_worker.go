package kafkaworker

import (
	"context"
	"fmt"
	"log"
	"strings"

	"agungdwiprasetyo.com/backend-microservices/config"
	"agungdwiprasetyo.com/backend-microservices/pkg/codebase/factory"
	"agungdwiprasetyo.com/backend-microservices/pkg/codebase/factory/constant"
	"agungdwiprasetyo.com/backend-microservices/pkg/codebase/interfaces"
	"agungdwiprasetyo.com/backend-microservices/pkg/helper"
	"github.com/Shopify/sarama"
)

type kafkaWorker struct {
	engine  sarama.ConsumerGroup
	service factory.ServiceFactory
}

// NewWorker create new HTTP server
func NewWorker(service factory.ServiceFactory) factory.AppServerFactory {
	// init kafka consumer
	kafkaConsumer, err := sarama.NewConsumerGroup(
		config.BaseEnv().Kafka.Brokers,
		config.BaseEnv().Kafka.ConsumerGroup,
		service.GetDependency().GetBroker().GetConfig(),
	)
	if err != nil {
		log.Panicf("Error creating kafka consumer group client: %v", err)
	}

	return &kafkaWorker{
		engine:  kafkaConsumer,
		service: service,
	}
}

func (h *kafkaWorker) Serve() {

	topicInfo := make(map[string][]string)
	var handlers = make(map[string][]interfaces.WorkerHandler)
	for _, m := range h.service.GetModules() {
		if h := m.WorkerHandler(constant.Kafka); h != nil {
			for _, topic := range h.GetTopics() {
				handlers[topic] = append(handlers[topic], h) // one same topic consumed by multiple module
				topicInfo[topic] = append(topicInfo[topic], string(m.Name()))
			}
		}
	}
	consumer := kafkaConsumer{
		handlers: handlers,
	}

	fmt.Println(helper.StringYellow("⇨ Kafka consumer is active"))
	var consumeTopics []string
	for topic, handlerNames := range topicInfo {
		fmt.Println(helper.StringYellow(fmt.Sprintf("[KAFKA-CONSUMER] (topic): %-8s  (consumed by modules)--> [%s]\n", topic, strings.Join(handlerNames, ", "))))
		consumeTopics = append(consumeTopics, topic)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := h.engine.Consume(ctx, consumeTopics, &consumer); err != nil {
		log.Panicf("Error from kafka consumer: %v", err)
	}
}

func (h *kafkaWorker) Shutdown(ctx context.Context) {
	log.Println("Stopping Kafka consumer...")
	h.engine.Close()
}

// kafkaConsumer represents a Sarama consumer group consumer
type kafkaConsumer struct {
	handlers map[string][]interfaces.WorkerHandler
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
			handler.ProcessMessage(session.Context(), message.Topic, message.Value)
		}

		session.MarkMessage(message, "")
	}

	return nil
}
