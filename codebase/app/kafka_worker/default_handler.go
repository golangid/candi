package kafkaworker

import (
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/IBM/sarama"
	"github.com/golangid/candi/broker"
	"github.com/golangid/candi/candishared"
	"github.com/golangid/candi/codebase/factory/types"
	"github.com/golangid/candi/tracer"
)

// consumerHandler represents a Sarama consumer group consumer
type consumerHandler struct {
	bk           *broker.KafkaBroker
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

	trace, ctx := tracer.StartTraceFromHeader(ctx, "KafkaConsumer", header)
	defer trace.Finish(
		tracer.FinishWithRecoverPanic(func(any) {}),
		tracer.FinishWithFunc(func() {
			if handler.AutoACK {
				session.MarkMessage(message, "")
			}
		}),
	)

	trace.SetTag("brokers", strings.Join(c.bk.BrokerHost, ","))
	trace.SetTag("topic", message.Topic)
	trace.SetTag("key", message.Key)
	trace.SetTag("consumer_group", c.opt.consumerGroup)
	if c.bk.WorkerType != types.Kafka {
		trace.SetTag("worker_type", string(c.bk.WorkerType))
	}
	trace.Log("header", header)
	trace.Log("message", message.Value)

	if c.opt.debugMode {
		log.Printf("\x1b[35;3mKafka Consumer%s: message consumed, timestamp = %v, topic = %s\x1b[0m", getWorkerTypeLog(c.bk.WorkerType), message.Timestamp, message.Topic)
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
		if err := handlerFunc(eventContext); err != nil {
			eventContext.SetError(err)
			trace.SetError(err)
		}
	}
}

func (c *consumerHandler) releaseMessagePool(eventContext *candishared.EventContext) {
	eventContext.Reset()
	c.messagePool.Put(eventContext)
}
