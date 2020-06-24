package usecase

import (
	"context"

	"agungdwiprasetyo.com/backend-microservices/internal/notification-service/modules/push-notif/domain"
)

// PushNotifUsecase abstraction
type PushNotifUsecase interface {
	SendNotification(ctx context.Context, request *domain.PushNotifRequestPayload) error
}
