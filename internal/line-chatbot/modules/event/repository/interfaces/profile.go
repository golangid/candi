package interfaces

import (
	"context"

	"agungdwiprasetyo.com/backend-microservices/internal/line-chatbot/modules/event/domain"
	"agungdwiprasetyo.com/backend-microservices/pkg/shared"
)

// Profile interface
type Profile interface {
	FindAll(ctx context.Context, filter *shared.Filter) <-chan shared.Result
	Count(ctx context.Context, filter *shared.Filter) <-chan int
	FindByID(context.Context, string) <-chan shared.Result
	Save(context.Context, *domain.Profile) <-chan error
}
