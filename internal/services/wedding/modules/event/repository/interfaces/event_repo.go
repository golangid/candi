package interfaces

import (
	"context"

	"agungdwiprasetyo.com/backend-microservices/internal/services/wedding/modules/event/domain"
	"agungdwiprasetyo.com/backend-microservices/pkg/shared"
)

// EventRepo abstraction
type EventRepo interface {
	Find(ctx context.Context, where domain.Event) <-chan shared.Result
}
