package publisher

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Shopify/sarama"
	"pkg.agungdwiprasetyo.com/candi/logger"
	"pkg.agungdwiprasetyo.com/candi/tracer"
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

	return &KafkaPublisher{producer}
}

// PublishMessage method
func (p *KafkaPublisher) PublishMessage(ctx context.Context, topic, key string, data interface{}) (err error) {
	opName := "kafka:publish_message"

	var payload []byte

	switch d := data.(type) {
	case string:
		payload = []byte(d)
	case []byte:
		payload = d
	default:
		payload, _ = json.Marshal(data)
	}

	tracer.WithTraceFunc(ctx, opName, func(ctx context.Context, tag map[string]interface{}) {
		// set tracer tag
		tag["topic"] = topic
		tag["key"] = key
		tag["message"] = string(payload)

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
	})

	return
}
