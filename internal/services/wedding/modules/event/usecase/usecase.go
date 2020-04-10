package usecase

import (
	"context"

	"agungdwiprasetyo.com/backend-microservices/internal/services/wedding/modules/event/domain"
)

// EventUsecase abstraction
type EventUsecase interface {
	FindByCode(ctx context.Context, code string) (*domain.Event, error)
}
