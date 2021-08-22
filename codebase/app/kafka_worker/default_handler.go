package kafkaworker

import (
	"fmt"
	"log"

	"github.com/Shopify/sarama"
	"pkg.agungdp.dev/candi/candishared"
	"pkg.agungdp.dev/candi/codebase/factory/types"
	"pkg.agungdp.dev/candi/logger"
	"pkg.agungdp.dev/candi/tracer"
)

// consumerHandler represents a Sarama consumer group consumer
type consumerHandler struct {
	opt          *option
	topics       []string
	handlerFuncs map[string]types.WorkerHandler
	ready        chan struct{}
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
	ctx := session.Context()
	handler := c.handlerFuncs[message.Topic]
	if handler.DisableTrace {
		ctx = tracer.SkipTraceContext(ctx)
	}
	trace, ctx := tracer.StartTraceWithContext(ctx, "KafkaConsumer")
	defer func() {
		if r := recover(); r != nil {
			trace.SetError(fmt.Errorf("%v", r))
		}

		if handler.AutoACK {
			session.MarkMessage(message, "")
		}
		logger.LogGreen("kafka_consumer > trace_url: " + tracer.GetTraceURL(ctx))
		trace.Finish()
	}()
	trace.SetTag("topic", message.Topic)
	trace.SetTag("key", string(message.Key))
	trace.SetTag("partition", message.Partition)
	trace.SetTag("offset", message.Offset)
	trace.SetTag("consumer_group", c.opt.consumerGroup)
	trace.Log("message", message.Value)

	if c.opt.debugMode {
		log.Printf("\x1b[35;3mKafka Consumer: message consumed, timestamp = %v, topic = %s\x1b[0m", message.Timestamp, message.Topic)
	}

	ctx = candishared.SetToContext(ctx, candishared.ContextKeyWorkerKey, message.Key)
	if err := handler.HandlerFunc(ctx, message.Value); err != nil {
		if handler.ErrorHandler != nil {
			handler.ErrorHandler(ctx, types.Kafka, message.Topic, message.Value, err)
		}
		trace.SetError(err)
	}
}
