package interfaces

import (
	"context"

	"pkg.agungdp.dev/candi/candishared"
)

// TokenValidator abstract interface
type TokenValidator interface {
	ValidateToken(ctx context.Context, token string) (*candishared.TokenClaim, error)
}
