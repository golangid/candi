package token

import (
	"context"
	"time"

	"agungdwiprasetyo.com/backend-microservices/pkg/shared"
)

// Token abstraction
type Token interface {
	Generate(ctx context.Context, payload *Claim, expired time.Duration) <-chan shared.Result
	Refresh(ctx context.Context, token string) <-chan shared.Result
	Validate(ctx context.Context, token string) <-chan shared.Result
	Revoke(ctx context.Context, token string) <-chan shared.Result
}
