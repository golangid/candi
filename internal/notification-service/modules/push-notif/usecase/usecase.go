package usecase

import "context"

// PushNotifUsecase abstraction
type PushNotifUsecase interface {
	SendNotification(ctx context.Context) error
}
