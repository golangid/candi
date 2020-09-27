package kafkaworker

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/Shopify/sarama"
	"pkg.agungdwiprasetyo.com/gendon/codebase/factory"
	"pkg.agungdwiprasetyo.com/gendon/codebase/factory/types"
	"pkg.agungdwiprasetyo.com/gendon/config"
	"pkg.agungdwiprasetyo.com/gendon/logger"
	"pkg.agungdwiprasetyo.com/gendon/tracer"
)

type kafkaWorker struct {
	engine          sarama.ConsumerGroup
	service         factory.ServiceFactory
	consumerHandler *kafkaConsumer
}

// NewWorker create new kafka consumer
func NewWorker(service factory.ServiceFactory) factory.AppServerFactory {
	// init kafka consumer
	consumerEngine, err := sarama.NewConsumerGroupFromClient(
		config.BaseEnv().Kafka.ConsumerGroup,
		service.GetDependency().GetBroker().GetClient(),
	)
	if err != nil {
		log.Panicf("Error creating kafka consumer group client: %v", err)
	}

	var consumerHandler kafkaConsumer
	consumerHandler.handlerFuncs = make(map[string]struct {
		handlerFunc   types.WorkerHandlerFunc
		errorHandlers []types.WorkerErrorHandler
	})
	for _, m := range service.GetModules() {
		if h := m.WorkerHandler(types.Kafka); h != nil {
			var handlerGroup types.WorkerHandlerGroup
			h.MountHandlers(&handlerGroup)
			for _, handler := range handlerGroup.Handlers {
				if _, ok := consumerHandler.handlerFuncs[handler.Pattern]; ok {
					logger.LogYellow(fmt.Sprintf("Kafka: warning, topic %s has been used in another module, overwrite handler func", handler.Pattern))
				}
				consumerHandler.handlerFuncs[handler.Pattern] = struct {
					handlerFunc   types.WorkerHandlerFunc
					errorHandlers []types.WorkerErrorHandler
				}{
					handlerFunc: handler.HandlerFunc, errorHandlers: handler.ErrorHandler,
				}
				consumerHandler.topics = append(consumerHandler.topics, handler.Pattern)
				logger.LogYellow(fmt.Sprintf("[KAFKA-CONSUMER] (topic): %-8s  (consumed by module)--> [%s]", handler.Pattern, m.Name()))
			}
		}
	}
	fmt.Printf("\x1b[34;1m⇨ Kafka consumer is active. Brokers: " + strings.Join(config.BaseEnv().Kafka.Brokers, ", ") + "\x1b[0m\n\n")

	return &kafkaWorker{
		engine:          consumerEngine,
		service:         service,
		consumerHandler: &consumerHandler,
	}
}

func (h *kafkaWorker) Serve() {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

startConsume:
	if err := h.engine.Consume(ctx, h.consumerHandler.topics, h.consumerHandler); err != nil {
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
	topics       []string
	handlerFuncs map[string]struct { // mapping topic to handler func in delivery layer
		handlerFunc   types.WorkerHandlerFunc
		errorHandlers []types.WorkerErrorHandler
	}
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
				}
				session.MarkMessage(message, "")
				logger.LogGreen(tracer.GetTraceURL(ctx))
			}()

			log.Printf("\x1b[35;3mKafka Consumer: message consumed, timestamp = %v, topic = %s\x1b[0m", message.Timestamp, message.Topic)

			tags["topic"] = message.Topic
			tags["key"] = string(message.Key)
			tags["value"] = string(message.Value)
			tags["partition"] = message.Partition
			tags["offset"] = message.Offset

			handler := c.handlerFuncs[message.Topic]
			if err := handler.handlerFunc(ctx, message.Value); err != nil {
				for _, errHandler := range handler.errorHandlers {
					errHandler(ctx, types.Kafka, message.Topic, message.Value, err)
				}
				panic(err)
			}
		})
	}

	return nil
}
