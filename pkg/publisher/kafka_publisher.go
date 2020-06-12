package publisher

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"agungdwiprasetyo.com/backend-microservices/config"
	"github.com/Shopify/sarama"
)

type kafkaPublisher struct {
	producer sarama.SyncProducer
}

// NewKafkaPublisher constructor
func NewKafkaPublisher(cfg *sarama.Config) Publisher {

	brokers := config.BaseEnv().Kafka.Brokers
	if len(brokers) == 0 || (len(brokers) == 1 && brokers[0] == "") {
		log.Printf("Kafka publisher: warning, missing kafka broker for publish message. Should be panicked when using kafka publisher")
		return nil
	}

	producer, err := sarama.NewSyncProducer(brokers, cfg)
	if err != nil {
		panic(fmt.Errorf("%v. Brokers: %v", err, brokers))
	}

	return &kafkaPublisher{producer}
}

// PublishMessage method
func (p *kafkaPublisher) PublishMessage(ctx context.Context, topic, key string, data interface{}) (err error) {
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
