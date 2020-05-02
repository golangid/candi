package interfaces

import (
	"context"

	"agungdwiprasetyo.com/backend-microservices/internal/services/cms/modules/public/domain"
	"agungdwiprasetyo.com/backend-microservices/pkg/shared"
)

// Visitor repo
type Visitor interface {
	Save(ctx context.Context, data *domain.Visitor) <-chan error
	FindAll(ctx context.Context, filter *shared.Filter) <-chan shared.Result
	Count(ctx context.Context, filter *shared.Filter) <-chan int
}
