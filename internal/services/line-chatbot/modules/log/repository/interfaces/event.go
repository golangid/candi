package interfaces

import (
	"context"

	"agungdwiprasetyo.com/backend-microservices/internal/services/line-chatbot/modules/log/domain"
	"agungdwiprasetyo.com/backend-microservices/pkg/shared"
)

// Event interface
type Event interface {
	FindAll(ctx context.Context, filter *shared.Filter) <-chan shared.Result
	Count(ctx context.Context, filter *shared.Filter) <-chan int
	Save(ctx context.Context, data *domain.Event) <-chan error
}
