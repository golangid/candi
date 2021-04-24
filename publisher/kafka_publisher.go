package publisher

import (
	"context"
	"fmt"
	"time"

	"github.com/Shopify/sarama"
	"pkg.agungdp.dev/candi/candihelper"
	"pkg.agungdp.dev/candi/candishared"
	"pkg.agungdp.dev/candi/logger"
	"pkg.agungdp.dev/candi/tracer"
)

// KafkaPublisher kafka
type KafkaPublisher struct {
	producer sarama.SyncProducer
}

// NewKafkaPublisher constructor
func NewKafkaPublisher(client sarama.Client) *KafkaPublisher {

	producer, err := sarama.NewSyncProducerFromClient(client)
	if err != nil {
		logger.LogYellow(fmt.Sprintf("(Kafka publisher: warning, %v. Should be panicked when using kafka publisher.) ", err))
		return nil
	}

	return &KafkaPublisher{
		producer: producer,
	}
}

// PublishMessage method
func (p *KafkaPublisher) PublishMessage(ctx context.Context, args *candishared.PublisherArgument) (err error) {
	trace := tracer.StartTrace(ctx, "kafka:publish_message")
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
		trace.SetError(err)
		trace.Finish()
	}()

	payload := candihelper.ToBytes(args.Data)

	trace.SetTag("topic", args.Topic)
	trace.SetTag("key", args.Key)
	trace.Log("message", payload)

	msg := &sarama.ProducerMessage{
		Topic:     args.Topic,
		Key:       sarama.ByteEncoder([]byte(args.Key)),
		Value:     sarama.ByteEncoder(payload),
		Timestamp: time.Now(),
	}
	_, _, err = p.producer.SendMessage(msg)
	return
}
