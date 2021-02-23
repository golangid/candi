package kafkaworker

import (
	"fmt"
	"log"

	"github.com/Shopify/sarama"
	"pkg.agungdp.dev/candi/candishared"
	"pkg.agungdp.dev/candi/codebase/factory/types"
	"pkg.agungdp.dev/candi/config/env"
	"pkg.agungdp.dev/candi/logger"
	"pkg.agungdp.dev/candi/tracer"
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
					logger.LogGreen("kafka_consumer > trace_url: " + tracer.GetTraceURL(ctx))
					trace.Finish()
					<-c.semaphore
				}()

				if env.BaseEnv().DebugMode {
					log.Printf("\x1b[35;3mKafka Consumer: message consumed, timestamp = %v, topic = %s\x1b[0m", message.Timestamp, message.Topic)
				}

				tags["topic"], tags["key"] = message.Topic, string(message.Key)
				tags["partition"], tags["offset"] = message.Partition, message.Offset
				tracer.Log(ctx, "message", message.Value)

				ctx = candishared.SetToContext(ctx, candishared.ContextKeyWorkerKey, message.Key)
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
