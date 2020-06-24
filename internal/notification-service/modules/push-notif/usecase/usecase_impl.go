package usecase

import (
	"context"

	"agungdwiprasetyo.com/backend-microservices/internal/notification-service/modules/push-notif/domain"
	"agungdwiprasetyo.com/backend-microservices/internal/notification-service/modules/push-notif/repository"
	"agungdwiprasetyo.com/backend-microservices/pkg/logger"
)

type pushNotifUsecaseImpl struct {
	repo *repository.Repository
}

// NewPushNotifUsecase constructor
func NewPushNotifUsecase(repo *repository.Repository) PushNotifUsecase {
	return &pushNotifUsecaseImpl{
		repo: repo,
	}
}

func (uc *pushNotifUsecaseImpl) SendNotification(ctx context.Context) (err error) {

	requestPayload := domain.PushRequest{
		Notification: &domain.Notification{
			Title:          "Testing",
			Body:           "Hello",
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
