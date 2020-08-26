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
	SendScheduledEvent(ctx context.Context, scheduledAt time.Time, request *domain.Event) (err error)

	PublishMessageToTopic(ctx context.Context, event *domain.Event) (*domain.Event, error)
	AddSubscriber(ctx context.Context, clientID, topic string) <-chan *domain.Event
}
