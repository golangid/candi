package publisher

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Shopify/sarama"
	"pkg.agungdp.dev/candi/tracer"
)

// KafkaPublisher kafka
type KafkaPublisher struct {
	producer sarama.SyncProducer
}

// PublishMessage method
func (p *KafkaPublisher) PublishMessage(ctx context.Context, topic, key string, data interface{}) (err error) {
	trace := tracer.StartTrace(ctx, "kafka:publish_message")
	defer func() {
		if r := recover(); r != nil {
			tracer.SetError(ctx, fmt.Errorf("%v", r))
		}
		trace.Finish()
	}()

	var payload []byte

	switch d := data.(type) {
	case string:
		payload = []byte(d)
	case []byte:
		payload = d
	default:
		payload, _ = json.Marshal(data)
	}

	// set tracer tag
	trace.SetTag("topic", topic)
	trace.SetTag("key", key)
	trace.SetTag("message", string(payload))

	msg := &sarama.ProducerMessage{
		Topic:     topic,
		Key:       sarama.ByteEncoder([]byte(key)),
		Value:     sarama.ByteEncoder(payload),
		Timestamp: time.Now(),
	}
	_, _, err = p.producer.SendMessage(msg)
	if err != nil {
		tracer.SetError(ctx, err)
	}

	return
}
