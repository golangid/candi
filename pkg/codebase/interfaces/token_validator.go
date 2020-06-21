package interfaces

import (
	"context"

	"agungdwiprasetyo.com/backend-microservices/pkg/shared"
)

// TokenValidator abstract interface
type TokenValidator interface {
	Validate(ctx context.Context, token string) <-chan shared.Result
}
