package interfaces

import (
	"context"

	"pkg.agungdwiprasetyo.com/gendon/shared"
)

// TokenValidator abstract interface
type TokenValidator interface {
	ValidateToken(ctx context.Context, token string) (*shared.TokenClaim, error)
}
