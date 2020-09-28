package interfaces

import (
	"context"

	"pkg.agungdwiprasetyo.com/candi/candishared"
)

// TokenValidator abstract interface
type TokenValidator interface {
	ValidateToken(ctx context.Context, token string) (*candishared.TokenClaim, error)
}
