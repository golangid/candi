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
	ready chan struct{}
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
		case message := <-claim.Messages():

			c.processMessage(session, message)

		case <-session.Context().Done():
			return nil

		}
	}
}

func (c *consumerHandler) processMessage(session sarama.ConsumerGroupSession, message *sarama.ConsumerMessage) {
	trace, ctx := tracer.StartTraceWithContext(session.Context(), "KafkaConsumer")
	defer func() {
		if r := recover(); r != nil {
			trace.SetError(fmt.Errorf("%v", r))
		}
		session.MarkMessage(message, "")
		logger.LogGreen("kafka_consumer > trace_url: " + tracer.GetTraceURL(ctx))
		trace.Finish()
	}()

	if env.BaseEnv().DebugMode {
		log.Printf("\x1b[35;3mKafka Consumer: message consumed, timestamp = %v, topic = %s\x1b[0m", message.Timestamp, message.Topic)
	}

	trace.SetTag("topic", message.Topic)
	trace.SetTag("key", string(message.Key))
	trace.SetTag("partition", message.Partition)
	trace.SetTag("offset", message.Offset)
	trace.Log("message", message.Value)

	ctx = candishared.SetToContext(ctx, candishared.ContextKeyWorkerKey, message.Key)
	handler := c.handlerFuncs[message.Topic]
	if err := handler.handlerFunc(ctx, message.Value); err != nil {
		for _, errHandler := range handler.errorHandlers {
			errHandler(ctx, types.Kafka, message.Topic, message.Value, err)
		}
		trace.SetError(err)
	}
}
