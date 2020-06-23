package workerhandler

import (
	"context"
	"fmt"

	"agungdwiprasetyo.com/backend-microservices/internal/notification-service/modules/push-notif/usecase"
	"agungdwiprasetyo.com/backend-microservices/pkg/logger"
)

// KafkaHandler struct
type KafkaHandler struct {
	topics []string
	uc     usecase.PushNotifUsecase
}

// NewKafkaHandler constructor
func NewKafkaHandler(uc usecase.PushNotifUsecase) *KafkaHandler {
	return &KafkaHandler{
		topics: []string{"push-notif"},
		uc:     uc,
	}
}

// GetTopics from kafka consumer
func (h *KafkaHandler) GetTopics() []string {
	return h.topics
}

// ProcessMessage from kafka consumer
func (h *KafkaHandler) ProcessMessage(ctx context.Context, topic string, message []byte) {
	fmt.Printf("message consumed by module push-notif. topic: %s, message: %s\n", topic, string(message))

	var err error
	switch topic {
	case "push-notif":
		err = h.uc.SendNotification(ctx)
	}

	if err != nil {
		logger.LogE(err.Error())
	}
}
