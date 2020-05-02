package interfaces

import (
	"context"

	"agungdwiprasetyo.com/backend-microservices/internal/services/cms/modules/public/domain"
)

// Visitor repo
type Visitor interface {
	Save(ctx context.Context, data *domain.Visitor) <-chan error
}
