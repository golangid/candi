package kafkaworker

import (
	"fmt"
	"log"

	"github.com/Shopify/sarama"
	"pkg.agungdwiprasetyo.com/candi/codebase/factory/types"
	"pkg.agungdwiprasetyo.com/candi/config/env"
	"pkg.agungdwiprasetyo.com/candi/logger"
	"pkg.agungdwiprasetyo.com/candi/tracer"
)

// consumerHandler represents a Sarama consumer group consumer
type consumerHandler struct {
	topics       []string
	handlerFuncs map[string]struct { // mapping topic to handler func in delivery layer
		handlerFunc   types.WorkerHandlerFunc
		errorHandlers []types.WorkerErrorHandler
	}
	ready     chan struct{}
	semaphore chan struct{} // for control maximum total goroutines when exec handlers
}

// Setup is run at the beginning of a new session, before ConsumeClaim
func (c *consumerHandler) Setup(sarama.ConsumerGroupSession) error {
	close(c.ready)
	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited
func (c *consumerHandler) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages().
func (c *consumerHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {

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
					logger.LogGreen("kafka consumer " + tracer.GetTraceURL(ctx))
					trace.Finish()
					<-c.semaphore
				}()

				if env.BaseEnv().DebugMode {
					log.Printf("\x1b[35;3mKafka Consumer: message consumed, timestamp = %v, topic = %s\x1b[0m", message.Timestamp, message.Topic)
				}

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
