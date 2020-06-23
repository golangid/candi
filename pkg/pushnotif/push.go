package pushnotif

import (
	"context"

	"agungdwiprasetyo.com/backend-microservices/pkg/shared"
)

// PushNotif abstraction
type PushNotif interface {
	Push(ctx context.Context, req PushRequest) <-chan shared.Result
}
