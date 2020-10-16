package kafkaworker

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/Shopify/sarama"
	"pkg.agungdwiprasetyo.com/candi/codebase/factory"
	"pkg.agungdwiprasetyo.com/candi/codebase/factory/types"
	"pkg.agungdwiprasetyo.com/candi/config"
	"pkg.agungdwiprasetyo.com/candi/logger"
	"pkg.agungdwiprasetyo.com/candi/tracer"
)

type kafkaWorker struct {
	engine          sarama.ConsumerGroup
	service         factory.ServiceFactory
	consumerHandler *kafkaConsumer
	cancelFunc      func()
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

	consumerHandler.semaphore = make(chan struct{}, config.BaseEnv().MaxGoroutines)
	return &kafkaWorker{
		engine:          consumerEngine,
		service:         service,
		consumerHandler: &consumerHandler,
	}
}

func (h *kafkaWorker) Serve() {

	ctx, cancel := context.WithCancel(context.Background())
	h.cancelFunc = cancel

startConsume:
	if err := h.engine.Consume(ctx, h.consumerHandler.topics, h.consumerHandler); err != nil {
		log.Printf("Error from kafka consumer: %v", err)
		goto startConsume
	}
}

func (h *kafkaWorker) Shutdown(ctx context.Context) {
	log.Println("Stopping Kafka Consumer worker...")
	defer func() { log.Println("Stopping Kafka Consumer: \x1b[32;1mSUCCESS\x1b[0m") }()

	h.cancelFunc()
	h.engine.Close()
}

// kafkaConsumer represents a Sarama consumer group consumer
type kafkaConsumer struct {
	topics       []string
	handlerFuncs map[string]struct { // mapping topic to handler func in delivery layer
		handlerFunc   types.WorkerHandlerFunc
		errorHandlers []types.WorkerErrorHandler
	}
	semaphore chan struct{} // for control maximum total goroutines when exec handlers
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

	for {
		select {
		case msg := <-claim.Messages():

			c.semaphore <- struct{}{}
			go func(message *sarama.ConsumerMessage) {
				trace := tracer.StartTrace(session.Context(), "KafkaConsumer")
				ctx, tags := trace.Context(), trace.Tags()
				defer func() {
					if r := recover(); r != nil {
						trace.SetError(fmt.Errorf("%v", r))
					}
					session.MarkMessage(message, "")
					logger.LogGreen(tracer.GetTraceURL(ctx))
					trace.Finish()
					<-c.semaphore
				}()

				log.Printf("\x1b[35;3mKafka Consumer: message consumed, timestamp = %v, topic = %s\x1b[0m", message.Timestamp, message.Topic)

				tags["topic"], tags["key"], tags["value"] = message.Topic, string(message.Key), string(message.Value)
				tags["partition"], tags["offset"] = message.Partition, message.Offset

				handler := c.handlerFuncs[message.Topic]
				if err := handler.handlerFunc(ctx, message.Value); err != nil {
					for _, errHandler := range handler.errorHandlers {
						errHandler(ctx, types.Kafka, message.Topic, message.Value, err)
					}
					trace.SetError(err)
				}
			}(msg)

		case <-session.Context().Done():
			return nil

		}
	}
}
