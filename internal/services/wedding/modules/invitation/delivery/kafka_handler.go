package delivery

import (
	"context"
	"fmt"
)

// KafkaHandler struct
type KafkaHandler struct {
	topics []string
}

// NewKafkaHandler constructor
func NewKafkaHandler(topics []string) *KafkaHandler {
	return &KafkaHandler{
		topics: topics,
	}
}

// GetTopics from kafka consumer
func (h *KafkaHandler) GetTopics() []string {
	return h.topics
}

// ProcessMessage from kafka consumer
func (h *KafkaHandler) ProcessMessage(ctx context.Context, message []byte) {
	fmt.Println("from invitation module", string(message))
}
