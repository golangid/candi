package kafkaworker

import (
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/Shopify/sarama"
	"github.com/golangid/candi/candishared"
	"github.com/golangid/candi/codebase/factory/types"
	"github.com/golangid/candi/logger"
	"github.com/golangid/candi/tracer"
)

// consumerHandler represents a Sarama consumer group consumer
type consumerHandler struct {
	opt          *option
	topics       []string
	handlerFuncs map[string]types.WorkerHandler
	ready        chan struct{}
	messagePool  sync.Pool
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
	handler, ok := c.handlerFuncs[message.Topic]
	if !ok {
		return
	}

	ctx := session.Context()
	if handler.DisableTrace {
		ctx = tracer.SkipTraceContext(ctx)
	}

	header := map[string]string{
		"offset":    strconv.Itoa(int(message.Offset)),
		"partition": strconv.Itoa(int(message.Partition)),
		"timestamp": message.Timestamp.Format(time.RFC3339),
	}
	for _, val := range message.Headers {
		header[string(val.Key)] = string(val.Value)
	}

	var err error
	trace, ctx := tracer.StartTraceFromHeader(ctx, "KafkaConsumer", header)
	defer func() {
		if r := recover(); r != nil {
			trace.SetTag("panic", true)
			err = fmt.Errorf("%v", r)
		}
		if handler.AutoACK {
			session.MarkMessage(message, "")
		}
		logger.LogGreen("kafka_consumer > trace_url: " + tracer.GetTraceURL(ctx))
		trace.SetTag("trace_id", tracer.GetTraceID(ctx))
		trace.Finish(tracer.FinishWithError(err))
	}()

	trace.SetTag("brokers", c.opt.brokers)
	trace.SetTag("topic", message.Topic)
	trace.SetTag("key", message.Key)
	trace.SetTag("consumer_group", c.opt.consumerGroup)
	trace.Log("header", header)
	trace.Log("message", message.Value)

	if c.opt.debugMode {
		log.Printf("\x1b[35;3mKafka Consumer: message consumed, timestamp = %v, topic = %s\x1b[0m", message.Timestamp, message.Topic)
	}

	eventContext := c.messagePool.Get().(*candishared.EventContext)
	defer c.releaseMessagePool(eventContext)
	eventContext.SetContext(ctx)
	eventContext.SetWorkerType(string(types.Kafka))
	eventContext.SetHandlerRoute(message.Topic)
	eventContext.SetHeader(header)
	eventContext.SetKey(string(message.Key))
	eventContext.Write(message.Value)

	for _, handlerFunc := range handler.HandlerFuncs {
		if err = handlerFunc(eventContext); err != nil {
			eventContext.SetError(err)
		}
	}
}

func (c *consumerHandler) releaseMessagePool(eventContext *candishared.EventContext) {
	eventContext.Reset()
	c.messagePool.Put(eventContext)
}
