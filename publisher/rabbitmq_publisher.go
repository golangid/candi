package publisher

import (
	"context"
	"fmt"
	"time"

	"github.com/streadway/amqp"
	"pkg.agungdp.dev/candi/candihelper"
	"pkg.agungdp.dev/candi/candishared"
	"pkg.agungdp.dev/candi/config/env"
	"pkg.agungdp.dev/candi/tracer"
)

const (
	// RabbitMQDelayHeader header key
	RabbitMQDelayHeader = "x-delay"
)

// RabbitMQPublisher rabbitmq
type RabbitMQPublisher struct {
	conn *amqp.Connection
}

// NewRabbitMQPublisher constructor
func NewRabbitMQPublisher(conn *amqp.Connection) *RabbitMQPublisher {
	return &RabbitMQPublisher{
		conn: conn,
	}
}

// PublishMessage method
func (r *RabbitMQPublisher) PublishMessage(ctx context.Context, args *candishared.PublisherArgument) (err error) {
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
		args.ContentType = "application/json"
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
