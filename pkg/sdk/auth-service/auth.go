package auth

import (
	"context"

	"agungdwiprasetyo.com/backend-microservices/pkg/shared"
)

// Auth service
type Auth interface {
	ValidateToken(ctx context.Context, token string) (cl *shared.TokenClaim, err error)
	GenerateToken(ctx context.Context, claim *shared.TokenClaim) <-chan shared.Result
}
