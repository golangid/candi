package interfaces

import (
	"context"

	"agungdwiprasetyo.com/backend-microservices/internal/notification-service/modules/push-notif/domain"
	"agungdwiprasetyo.com/backend-microservices/pkg/shared"
)

// PushNotif abstraction
type PushNotif interface {
	Push(ctx context.Context, req domain.PushRequest) <-chan shared.Result
}
