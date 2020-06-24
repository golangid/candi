package workerhandler

import (
	"context"
	"encoding/json"
	"fmt"

	"agungdwiprasetyo.com/backend-microservices/internal/notification-service/modules/push-notif/domain"
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
		var payload domain.PushNotifRequestPayload
		json.Unmarshal(message, &payload)
		err = h.uc.SendNotification(ctx, &payload)
	}

	if err != nil {
		logger.LogE(err.Error())
	}
}
