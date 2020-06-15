package usecase

import (
	"context"

	"agungdwiprasetyo.com/backend-microservices/internal/line-chatbot/modules/event/domain"
	"agungdwiprasetyo.com/backend-microservices/pkg/shared"
)

// EventUsecase abstraction
type EventUsecase interface {
	FindAll(ctx context.Context, filter *shared.Filter) (data []domain.Event, meta *shared.Meta, err error)
}
