package middleware

import (
	"context"
	"errors"
	"strings"

	"github.com/golangid/candi/candishared"
	"github.com/golangid/candi/codebase/interfaces"
	"github.com/golangid/candi/config/env"
)

// Middleware impl
type Middleware struct {
	TokenValidator       interfaces.TokenValidator
	ACLPermissionChecker interfaces.ACLPermissionChecker
	authTypeCheckerFunc  map[string]func(context.Context, string) (*candishared.TokenClaim, error)
}

// NewMiddleware create new middleware instance
func NewMiddleware(tokenValidator interfaces.TokenValidator, aclPermissionChecker interfaces.ACLPermissionChecker) *Middleware {
	mw := &Middleware{
		TokenValidator: tokenValidator, ACLPermissionChecker: aclPermissionChecker,
	}

	mw.authTypeCheckerFunc = map[string]func(context.Context, string) (*candishared.TokenClaim, error){
		Basic: func(ctx context.Context, key string) (*candishared.TokenClaim, error) {
			return nil, mw.Basic(ctx, key)
		},
		Bearer: func(ctx context.Context, token string) (*candishared.TokenClaim, error) {
			return mw.Bearer(ctx, token)
		},
	}

	return mw
}

func extractAuthType(prefix, authorization string) (string, error) {
	if env.BaseEnv().NoAuth {
		return "", nil
	}

	authValues := strings.Split(authorization, " ")
	if len(authValues) == 2 && strings.ToLower(authValues[0]) == prefix {
		return authValues[1], nil
	}

	return "", errors.New("Invalid authorization")
}
