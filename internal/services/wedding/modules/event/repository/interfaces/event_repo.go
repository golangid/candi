package interfaces

import (
	"context"

	"github.com/agungdwiprasetyo/backend-microservices/internal/services/wedding/modules/event/domain"
	"github.com/agungdwiprasetyo/backend-microservices/pkg/shared"
)

// EventRepo abstraction
type EventRepo interface {
	Find(ctx context.Context, where domain.Event) <-chan shared.Result
}
