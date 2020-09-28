package interfaces

import (
	"context"

	"pkg.agungdwiprasetyo.com/candi/shared"
)

// TokenValidator abstract interface
type TokenValidator interface {
	ValidateToken(ctx context.Context, token string) (*shared.TokenClaim, error)
}
