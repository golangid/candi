package workerhandler

import (
	"context"
	"fmt"
)

// KafkaHandler struct
type KafkaHandler struct {
	topics []string
}

// NewKafkaHandler constructor
func NewKafkaHandler() *KafkaHandler {
	return &KafkaHandler{
		topics: []string{"test", "update-member"},
	}
}

// GetTopics from kafka consumer
func (h *KafkaHandler) GetTopics() []string {
	return h.topics
}

// ProcessMessage from kafka consumer
func (h *KafkaHandler) ProcessMessage(ctx context.Context, topic string, message []byte) {
	fmt.Printf("message consumed by module auth. message: %s\n", string(message))
}
