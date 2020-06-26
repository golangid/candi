package usecase

import (
	"context"
	"encoding/json"
	"time"

	"agungdwiprasetyo.com/backend-microservices/internal/notification-service/modules/push-notif/domain"
	"agungdwiprasetyo.com/backend-microservices/internal/notification-service/modules/push-notif/repository"
	"agungdwiprasetyo.com/backend-microservices/pkg/codebase/factory/constant"
	"agungdwiprasetyo.com/backend-microservices/pkg/helper"
	"agungdwiprasetyo.com/backend-microservices/pkg/logger"
	"agungdwiprasetyo.com/backend-microservices/pkg/utils"
)

type pushNotifUsecaseImpl struct {
	modName constant.Module
	repo    *repository.Repository
}

// NewPushNotifUsecase constructor
func NewPushNotifUsecase(modName constant.Module, repo *repository.Repository) PushNotifUsecase {
	return &pushNotifUsecaseImpl{
		modName: modName,
		repo:    repo,
	}
}

func (uc *pushNotifUsecaseImpl) SendNotification(ctx context.Context, request *domain.PushNotifRequestPayload) (err error) {
	trace := utils.StartTrace(ctx, "Usecase-SendNotification")
	defer trace.Finish()

	requestPayload := domain.PushRequest{
		To: request.To,
		Notification: &domain.Notification{
			Title:          request.Title,
			Body:           request.Message,
			Image:          "https://storage.googleapis.com/agungdp/static/logo/golang.png",
			Sound:          "default",
			MutableContent: true,
			ResourceID:     "resourceID",
			ResourceName:   "resourceName",
		},
		Data: map[string]interface{}{"type": "type"},
	}
	result := <-uc.repo.PushNotif.Push(ctx, requestPayload)
	if result.Error != nil {
		return result.Error
	}

	logger.LogI("success send notification")
	return
}

func (uc *pushNotifUsecaseImpl) SendScheduledNotification(ctx context.Context, scheduledAt time.Time, request *domain.PushNotifRequestPayload) (err error) {

	redisTopicKey := helper.BuildRedisPubSubKeyTopic(string(uc.modName), "scheduled-push-notif")
	data, _ := json.Marshal(request)
	exp := scheduledAt.Sub(time.Now())
	return uc.repo.Schedule.SaveScheduledNotification(ctx, redisTopicKey, data, exp)
}
