package workerhandler

import (
	"context"
	"fmt"

	"agungdwiprasetyo.com/backend-microservices/pkg/codebase/factory/types"
)

// KafkaHandler struct
type KafkaHandler struct {
}

// NewKafkaHandler constructor
func NewKafkaHandler() *KafkaHandler {
	return &KafkaHandler{}
}

// MountHandlers return map topic to handler func
func (h *KafkaHandler) MountHandlers() map[string]types.WorkerHandlerFunc {

	return map[string]types.WorkerHandlerFunc{
		"test": h.handleTest,
	}
}

// ProcessMessage from kafka consumer
func (h *KafkaHandler) handleTest(ctx context.Context, message []byte) error {
	fmt.Printf("message consumed by module token. message: %s\n", string(message))
	return nil
}
