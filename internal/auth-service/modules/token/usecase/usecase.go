package usecase

import (
	"context"

	"agungdwiprasetyo.com/backend-microservices/pkg/shared"
)

// TokenUsecase abstraction
type TokenUsecase interface {
	Generate(ctx context.Context, payload *shared.TokenClaim) <-chan shared.Result
	Refresh(ctx context.Context, token string) <-chan shared.Result
	Validate(ctx context.Context, token string) <-chan shared.Result
	Revoke(ctx context.Context, token string) <-chan shared.Result
}
