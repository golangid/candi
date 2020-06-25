package workerhandler

import (
	"context"
	"fmt"

	"agungdwiprasetyo.com/backend-microservices/internal/notification-service/modules/push-notif/usecase"
	"agungdwiprasetyo.com/backend-microservices/pkg/codebase/factory/constant"
	"agungdwiprasetyo.com/backend-microservices/pkg/helper"
	"agungdwiprasetyo.com/backend-microservices/pkg/logger"
)

// RedisHandler struct
type RedisHandler struct {
	topics []string
	uc     usecase.PushNotifUsecase
}

// NewRedisHandler constructor
func NewRedisHandler(modName constant.Module, uc usecase.PushNotifUsecase) *RedisHandler {
	return &RedisHandler{
		topics: []string{
			helper.BuildRedisPubSubKeyTopic(string(modName), "push-notif"),
			helper.BuildRedisPubSubKeyTopic(string(modName), "push"),
		},
		uc: uc,
	}
}

// GetTopics from redis worker
func (h *RedisHandler) GetTopics() []string {
	return h.topics
}

// ProcessMessage from redis worker
func (h *RedisHandler) ProcessMessage(ctx context.Context, topic string, message []byte) {
	logger.LogIf("PushNotif module: redis subscriber run on topic: %s, message: %s", topic, string(message))

	var err error
	switch topic {
	case "push-notif":
		fmt.Println("mantab")
	case "push":
		fmt.Println("wkwkwk")
	}

	if err != nil {
		logger.LogE(err.Error())
	}
}
