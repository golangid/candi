package broker

import (
	"context"
	"fmt"
	"time"

	"github.com/golangid/candi/candihelper"
	"github.com/golangid/candi/candishared"
	"github.com/golangid/candi/codebase/factory/types"
	"github.com/golangid/candi/codebase/interfaces"
	"github.com/golangid/candi/config/env"
	"github.com/golangid/candi/logger"
	"github.com/golangid/candi/tracer"
	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	// RabbitMQDelayHeader header key, value in millisecond
	RabbitMQDelayHeader = "x-delay"
)

// RabbitMQOptionFunc func type
type RabbitMQOptionFunc func(*RabbitMQBroker)

// RabbitMQSetWorkerType set worker type
func RabbitMQSetWorkerType(workerType types.Worker) RabbitMQOptionFunc {
	return func(bk *RabbitMQBroker) {
		bk.WorkerType = workerType
	}
}

// RabbitMQSetBrokerHost set custom broker host
func RabbitMQSetBrokerHost(brokers string) RabbitMQOptionFunc {
	return func(bk *RabbitMQBroker) {
		bk.BrokerHost = brokers
	}
}

// RabbitMQSetExchange set exchange
func RabbitMQSetExchange(exchange string) RabbitMQOptionFunc {
	return func(bk *RabbitMQBroker) {
		bk.Exchange = exchange
	}
}

// RabbitMQSetChannel set custom channel configuration
func RabbitMQSetChannel(ch *amqp.Channel) RabbitMQOptionFunc {
	return func(bk *RabbitMQBroker) {
		bk.Channel = ch
	}
}

// RabbitMQSetPublisher set custom publisher
func RabbitMQSetPublisher(pub interfaces.Publisher) RabbitMQOptionFunc {
	return func(bk *RabbitMQBroker) {
		bk.publisher = pub
	}
}

// RabbitMQBroker broker
type RabbitMQBroker struct {
	publisher interfaces.Publisher

	WorkerType types.Worker
	BrokerHost string
	Exchange   string
	Conn       *amqp.Connection
	Channel    *amqp.Channel
}

// NewRabbitMQBroker setup rabbitmq configuration for publisher or consumer, default connection from RABBITMQ_BROKER environment (with default worker type is types.RabbitMQ)
func NewRabbitMQBroker(opts ...RabbitMQOptionFunc) *RabbitMQBroker {
	defer logger.LogWithDefer("Load RabbitMQ broker configuration... ")()
	var err error

	bk := new(RabbitMQBroker)
	bk.BrokerHost = env.BaseEnv().RabbitMQ.Broker
	bk.Exchange = env.BaseEnv().RabbitMQ.ExchangeName
	bk.WorkerType = types.RabbitMQ
	for _, opt := range opts {
		opt(bk)
	}

	bk.Conn, err = amqp.Dial(bk.BrokerHost)
	if err != nil {
		panic("RabbitMQ: cannot connect to server broker: " + err.Error())
	}

	if bk.Channel == nil {
		// set default channel configuration
		bk.Channel, err = bk.Conn.Channel()
		if err != nil {
			panic("RabbitMQ channel: " + err.Error())
		}
		if err := bk.Channel.ExchangeDeclare("amq.direct", "direct", true, false, false, false, nil); err != nil {
			panic("RabbitMQ exchange declare direct: " + err.Error())
		}
		if err := bk.Channel.ExchangeDeclare(
			bk.Exchange,         // name
			"x-delayed-message", // type
			true,                // durable
			false,               // auto-deleted
			false,               // internal
			false,               // no-wait
			amqp.Table{
				"x-delayed-type": "direct",
			},
		); err != nil {
			panic("RabbitMQ exchange declare delayed: " + err.Error())
		}
		if err := bk.Channel.Qos(2, 0, false); err != nil {
			panic("RabbitMQ Qos: " + err.Error())
		}
	}

	if bk.publisher == nil {
		bk.publisher = NewRabbitMQPublisher(bk.Conn, bk.Exchange)
	}

	return bk
}

// GetPublisher method
func (r *RabbitMQBroker) GetPublisher() interfaces.Publisher {
	return r.publisher
}

// GetName method
func (r *RabbitMQBroker) GetName() types.Worker {
	return r.WorkerType
}

// Health method
func (r *RabbitMQBroker) Health() map[string]error {
	return map[string]error{string(types.RabbitMQ): nil}
}

// Disconnect method
func (r *RabbitMQBroker) Disconnect(ctx context.Context) error {
	defer logger.LogWithDefer("\x1b[33;5mrabbitmq_broker\x1b[0m: disconnect...")()

	return r.Conn.Close()
}

// RabbitMQPublisher rabbitmq
type RabbitMQPublisher struct {
	conn     *amqp.Connection
	exchange string
}

// NewRabbitMQPublisher setup only rabbitmq publisher with client connection
func NewRabbitMQPublisher(conn *amqp.Connection, exchange string) *RabbitMQPublisher {
	return &RabbitMQPublisher{
		conn: conn, exchange: exchange,
	}
}

// PublishMessage method
func (r *RabbitMQPublisher) PublishMessage(ctx context.Context, args *candishared.PublisherArgument) (err error) {
	trace, _ := tracer.StartTraceWithContext(ctx, "rabbitmq:publish_message")
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
		trace.Finish(tracer.FinishWithError(err))
	}()

	ch, err := r.conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	if args.ContentType == "" {
		args.ContentType = candihelper.HeaderMIMEApplicationJSON
	}

	if args.Header == nil {
		args.Header = make(map[string]any)
	}

	traceHeader := map[string]string{}
	trace.InjectRequestHeader(traceHeader)
	for k, v := range traceHeader {
		args.Header[k] = v
	}

	trace.SetTag("topic", args.Topic)
	trace.SetTag("key", args.Key)

	msg := amqp.Publishing{
		DeliveryMode: amqp.Persistent,
		Timestamp:    time.Now(),
		ContentType:  args.ContentType,
		Headers:      amqp.Table(args.Header),
	}
	if args.Delay > 0 {
		msg.Headers[RabbitMQDelayHeader] = args.Delay.Milliseconds()
	}

	if len(args.Message) > 0 {
		msg.Body = args.Message
	} else {
		msg.Body = candihelper.ToBytes(args.Data)
	}

	trace.Log("header", msg.Headers)
	trace.Log("message", msg.Body)

	return ch.PublishWithContext(ctx,
		r.exchange,
		args.Topic, // routing key
		false,      // mandatory
		false,      // immediate
		msg)
}
