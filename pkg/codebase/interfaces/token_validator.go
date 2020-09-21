package interfaces

import (
	"context"

	"agungdwiprasetyo.com/backend-microservices/pkg/shared"
)

// TokenValidator abstract interface
type TokenValidator interface {
	ValidateToken(ctx context.Context, token string) (*shared.TokenClaim, error)
}
