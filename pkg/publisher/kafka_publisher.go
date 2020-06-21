package publisher

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"agungdwiprasetyo.com/backend-microservices/config"
	"github.com/Shopify/sarama"
)

// KafkaPublisher kafka
type KafkaPublisher struct {
	producer sarama.SyncProducer
}

// NewKafkaPublisher constructor
func NewKafkaPublisher(cfg *sarama.Config) *KafkaPublisher {

	brokers := config.BaseEnv().Kafka.Brokers
	if len(brokers) == 0 || (len(brokers) == 1 && brokers[0] == "") {
		log.Printf("Kafka publisher: warning, missing kafka broker for publish message. Should be panicked when using kafka publisher")
		return nil
	}

	producer, err := sarama.NewSyncProducer(brokers, cfg)
	if err != nil {
		log.Printf("Kafka publisher: warning, %v. Should be panicked when using kafka publisher", err)
		return nil
	}

	return &KafkaPublisher{producer}
}

// PublishMessage method
func (p *KafkaPublisher) PublishMessage(ctx context.Context, topic, key string, data interface{}) (err error) {
	payload, _ := json.Marshal(data)

	msg := &sarama.ProducerMessage{
		Topic:     topic,
		Key:       sarama.ByteEncoder([]byte(key)),
		Value:     sarama.ByteEncoder(payload),
		Timestamp: time.Now(),
	}
	_, _, err = p.producer.SendMessage(msg)

	return
}
