package workerhandler

import (
	"context"
	"fmt"
	"time"

	"agungdwiprasetyo.com/backend-microservices/internal/notification-service/modules/push-notif/usecase"
	"agungdwiprasetyo.com/backend-microservices/pkg/helper"
	"agungdwiprasetyo.com/backend-microservices/pkg/logger"
)

// CronHandler struct
type CronHandler struct {
	topics []string
	uc     usecase.PushNotifUsecase
}

// NewCronHandler constructor
func NewCronHandler(uc usecase.PushNotifUsecase) *CronHandler {
	return &CronHandler{
		topics: []string{
			helper.CronJobKeyToString("push-notif", "23:30:17"),
			helper.CronJobKeyToString("push", "2s"),
		},
		uc: uc,
	}
}

// GetTopics from cron worker
func (h *CronHandler) GetTopics() []string {
	return h.topics
}

// ProcessMessage from cron worker
func (h *CronHandler) ProcessMessage(ctx context.Context, topic string, message []byte) {
	logger.LogIf("PushNotif module: scheduler run on topic: %s, message: %s\n", topic, string(message))

	var err error
	switch topic {
	case "push-notif":
		fmt.Println("mantab")
	case "push":
		fmt.Println("wkwkwk")
		time.Sleep(50 * time.Second)
		fmt.Println("wkwkwk done")
	}

	if err != nil {
		logger.LogE(err.Error())
	}
}
