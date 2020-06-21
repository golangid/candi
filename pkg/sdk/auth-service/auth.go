package auth

import (
	"context"

	"agungdwiprasetyo.com/backend-microservices/pkg/shared"
)

// Auth service
type Auth interface {
	Validate(ctx context.Context, token string) <-chan shared.Result
}
