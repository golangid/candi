package middleware

import (
	"errors"
	"strings"

	"github.com/golangid/candi/codebase/interfaces"
	"github.com/golangid/candi/config/env"
)

// Middleware impl
type Middleware struct {
	TokenValidator       interfaces.TokenValidator
	ACLPermissionChecker interfaces.ACLPermissionChecker
}

// NewMiddleware create new middleware instance
func NewMiddleware(tokenValidator interfaces.TokenValidator, aclPermissionChecker interfaces.ACLPermissionChecker) *Middleware {
	mw := &Middleware{
		TokenValidator: tokenValidator, ACLPermissionChecker: aclPermissionChecker,
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
