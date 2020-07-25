package usecase

import (
	"context"
	"time"

	"agungdwiprasetyo.com/backend-microservices/internal/notification-service/modules/push-notif/domain"
)

// PushNotifUsecase abstraction
type PushNotifUsecase interface {
	SendNotification(ctx context.Context, request *domain.PushNotifRequestPayload) error
	SendScheduledNotification(ctx context.Context, scheduledAt time.Time, request *domain.PushNotifRequestPayload) (err error)

	SayHello(ctx context.Context, event *domain.HelloSaidEvent) *domain.HelloSaidEvent
	AddSubscriber(ctx context.Context) <-chan *domain.HelloSaidEvent
}
