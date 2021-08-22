package broker

import (
	"context"
	"fmt"
	"time"

	"github.com/streadway/amqp"
	"pkg.agungdp.dev/candi/candihelper"
	"pkg.agungdp.dev/candi/candishared"
	"pkg.agungdp.dev/candi/codebase/factory/types"
	"pkg.agungdp.dev/candi/codebase/interfaces"
	"pkg.agungdp.dev/candi/config/env"
	"pkg.agungdp.dev/candi/logger"
	"pkg.agungdp.dev/candi/tracer"
)

const (
	// RabbitMQDelayHeader header key, value in millisecond
	RabbitMQDelayHeader = "x-delay"
)

// RabbitMQOptionFunc func type
type RabbitMQOptionFunc func(*RabbitMQBroker)

// RabbitMQSetBrokerHost set custom broker host
func RabbitMQSetBrokerHost(brokers string) RabbitMQOptionFunc {
	return func(bk *RabbitMQBroker) {
		bk.brokerHost = brokers
	}
}

// RabbitMQSetChannel set custom channel configuration
func RabbitMQSetChannel(ch *amqp.Channel) RabbitMQOptionFunc {
	return func(bk *RabbitMQBroker) {
		bk.ch = ch
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
	brokerHost string
	conn       *amqp.Connection
	ch         *amqp.Channel
	publisher  interfaces.Publisher
}

// NewRabbitMQBroker setup rabbitmq configuration for publisher or consumer, default connection from RABBITMQ_BROKER environment
func NewRabbitMQBroker(opts ...RabbitMQOptionFunc) *RabbitMQBroker {
	deferFunc := logger.LogWithDefer("Load RabbitMQ broker configuration... ")
	defer deferFunc()
	var err error

	rabbitmq := new(RabbitMQBroker)
	rabbitmq.brokerHost = env.BaseEnv().RabbitMQ.Broker
	for _, opt := range opts {
		opt(rabbitmq)
	}

	rabbitmq.conn, err = amqp.Dial(rabbitmq.brokerHost)
	if err != nil {
		panic("RabbitMQ: cannot connect to server broker: " + err.Error())
	}

	if rabbitmq.ch == nil {
		// set default configuration
		rabbitmq.ch, err = rabbitmq.conn.Channel()
		if err != nil {
			panic("RabbitMQ channel: " + err.Error())
		}
		if err := rabbitmq.ch.ExchangeDeclare("amq.direct", "direct", true, false, false, false, nil); err != nil {
			panic("RabbitMQ exchange declare direct: " + err.Error())
		}
		if err := rabbitmq.ch.ExchangeDeclare(
			env.BaseEnv().RabbitMQ.ExchangeName, // name
			"x-delayed-message",                 // type
			true,                                // durable
			false,                               // auto-deleted
			false,                               // internal
			false,                               // no-wait
			amqp.Table{
				"x-delayed-type": "direct",
			},
		); err != nil {
			panic("RabbitMQ exchange declare delayed: " + err.Error())
		}
		if err := rabbitmq.ch.Qos(2, 0, false); err != nil {
			panic("RabbitMQ Qos: " + err.Error())
		}
	}

	if rabbitmq.publisher == nil {
		rabbitmq.publisher = NewRabbitMQPublisher(rabbitmq.conn)
	}

	return rabbitmq
}

// GetConfiguration method
func (r *RabbitMQBroker) GetConfiguration() interface{} {
	return r.ch
}

// GetPublisher method
func (r *RabbitMQBroker) GetPublisher() interfaces.Publisher {
	return r.publisher
}

// GetName method
func (r *RabbitMQBroker) GetName() types.Worker {
	return types.RabbitMQ
}

// Health method
func (r *RabbitMQBroker) Health() map[string]error {
	return map[string]error{string(types.RabbitMQ): nil}
}

// Disconnect method
func (r *RabbitMQBroker) Disconnect(ctx context.Context) error {
	deferFunc := logger.LogWithDefer("rabbitmq: disconnect...")
	defer deferFunc()

	return r.conn.Close()
}

// rabbitMQPublisher rabbitmq
type rabbitMQPublisher struct {
	conn *amqp.Connection
}

// NewRabbitMQPublisher setup only rabbitmq publisher with client connection
func NewRabbitMQPublisher(conn *amqp.Connection) interfaces.Publisher {
	return &rabbitMQPublisher{
		conn: conn,
	}
}

// PublishMessage method
func (r *rabbitMQPublisher) PublishMessage(ctx context.Context, args *candishared.PublisherArgument) (err error) {
	trace := tracer.StartTrace(ctx, "rabbitmq:publish_message")
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
		trace.SetError(err)
		trace.Finish()
	}()

	ch, err := r.conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	if args.ContentType == "" {
		args.ContentType = candihelper.HeaderMIMEApplicationJSON
	}

	trace.SetTag("topic", args.Topic)
	trace.SetTag("key", args.Key)

	msg := amqp.Publishing{
		DeliveryMode: amqp.Persistent,
		Timestamp:    time.Now(),
		ContentType:  args.ContentType,
		Body:         candihelper.ToBytes(args.Data),
		Headers:      amqp.Table(args.Header),
	}

	trace.Log("header", msg.Headers)
	trace.Log("message", msg.Body)

	return ch.Publish(
		env.BaseEnv().RabbitMQ.ExchangeName,
		args.Topic, // routing key
		false,      // mandatory
		false,      // immediate
		msg)
}
