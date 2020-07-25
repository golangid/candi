package kafkaworker

import (
	"context"
	"fmt"
	"log"
	"strings"

	"agungdwiprasetyo.com/backend-microservices/config"
	"agungdwiprasetyo.com/backend-microservices/pkg/codebase/factory"
	"agungdwiprasetyo.com/backend-microservices/pkg/codebase/factory/types"
	"agungdwiprasetyo.com/backend-microservices/pkg/logger"
	"agungdwiprasetyo.com/backend-microservices/pkg/tracer"
	"github.com/Shopify/sarama"
)

type kafkaWorker struct {
	engine  sarama.ConsumerGroup
	service factory.ServiceFactory
}

// NewWorker create new kafka consumer
func NewWorker(service factory.ServiceFactory) factory.AppServerFactory {
	// init kafka consumer
	kafkaConsumer, err := sarama.NewConsumerGroupFromClient(
		config.BaseEnv().Kafka.ConsumerGroup,
		service.GetDependency().GetBroker().GetClient(),
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

	var consumeTopics []string
	handlerFuncs := make(map[string]types.WorkerHandlerFunc)
	for _, m := range h.service.GetModules() {
		if h := m.WorkerHandler(types.Kafka); h != nil {
			for topic, handlerFunc := range h.MountHandlers() {
				if _, ok := handlerFuncs[topic]; ok {
					logger.LogYellow(fmt.Sprintf("Kafka: warning, topic %s has been used in another module, overwrite handler func", topic))
				}
				handlerFuncs[topic] = handlerFunc
				consumeTopics = append(consumeTopics, topic)
				logger.LogYellow(fmt.Sprintf("[KAFKA-CONSUMER] (topic): %-8s  (consumed by module)--> [%s]", topic, m.Name()))
			}
		}
	}

	consumer := kafkaConsumer{
		handlerFuncs: handlerFuncs,
	}

	fmt.Printf("\x1b[34;1m⇨ Kafka consumer is active. Brokers: " + strings.Join(config.BaseEnv().Kafka.Brokers, ", ") + "\x1b[0m\n\n")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

startConsume:
	if err := h.engine.Consume(ctx, consumeTopics, &consumer); err != nil {
		log.Printf("Error from kafka consumer: %v", err)
		goto startConsume
	}
}

func (h *kafkaWorker) Shutdown(ctx context.Context) {
	deferFunc := logger.LogWithDefer("Stopping Kafka consumer worker...")
	defer deferFunc()

	h.engine.Close()
}

// kafkaConsumer represents a Sarama consumer group consumer
type kafkaConsumer struct {
	handlerFuncs map[string]types.WorkerHandlerFunc // mapping topic to handler func in delivery layer
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

		tracer.WithTraceFunc(session.Context(), "KafkaConsumer", func(ctx context.Context, tags map[string]interface{}) {
			defer func() {
				if r := recover(); r != nil {
					tracer.SetError(ctx, fmt.Errorf("%v", r))
				} else {
					session.MarkMessage(message, "")
				}
				logger.LogGreen(tracer.GetTraceURL(ctx))
			}()

			log.Printf("Message claimed: timestamp = %v, topic = %s", message.Timestamp, message.Topic)

			tags["topic"] = message.Topic
			tags["key"] = string(message.Key)
			tags["value"] = string(message.Value)
			tags["partition"] = message.Partition
			tags["offset"] = message.Offset

			handlerFunc := c.handlerFuncs[message.Topic]
			if err := handlerFunc(ctx, message.Value); err != nil {
				panic(err)
			}
		})
	}

	return nil
}
