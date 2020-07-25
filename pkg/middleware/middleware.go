package middleware

import (
	"context"

	"agungdwiprasetyo.com/backend-microservices/config"
	"agungdwiprasetyo.com/backend-microservices/pkg/codebase/interfaces"
	"agungdwiprasetyo.com/backend-microservices/pkg/shared"
)

// Middleware impl
type Middleware struct {
	tokenValidator      interfaces.TokenValidator
	username, password  string
	authTypeCheckerFunc map[string]func(context.Context, string) (*shared.TokenClaim, error)
}

// NewMiddleware create new middleware instance
func NewMiddleware(tokenValidator interfaces.TokenValidator) *Middleware {
	mw := &Middleware{
		tokenValidator: tokenValidator,
		username:       config.BaseEnv().BasicAuthUsername,
		password:       config.BaseEnv().BasicAuthPassword,
	}

	mw.authTypeCheckerFunc = map[string]func(context.Context, string) (*shared.TokenClaim, error){
		Basic: func(ctx context.Context, key string) (*shared.TokenClaim, error) {
			return nil, mw.Basic(ctx, key)
		},
		Bearer: func(ctx context.Context, token string) (*shared.TokenClaim, error) {
			return mw.Bearer(ctx, token)
		},
	}

	return mw
}
