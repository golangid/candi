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
	BasicAuthValidator   interfaces.BasicAuthValidator
}

// NewMiddleware create new middleware instance
func NewMiddleware(tokenValidator interfaces.TokenValidator, aclPermissionChecker interfaces.ACLPermissionChecker) *Middleware {
	mw := &Middleware{
		TokenValidator: tokenValidator, ACLPermissionChecker: aclPermissionChecker,
		BasicAuthValidator: &defaultMiddleware{},
	}

	return mw
}

// NewMiddlewareWithOption create new middleware instance with option
func NewMiddlewareWithOption(opts ...OptionFunc) *Middleware {
	defaultMw := &defaultMiddleware{}
	mw := &Middleware{
		TokenValidator: defaultMw, ACLPermissionChecker: defaultMw, BasicAuthValidator: defaultMw,
	}
	for _, opt := range opts {
		opt(mw)
	}

	return mw
}

func extractAuthType(prefix, authorization string) (string, error) {

	authValues := strings.Split(authorization, " ")
	if len(authValues) == 2 && strings.ToLower(authValues[0]) == prefix {
		return authValues[1], nil
	}

	return "", errors.New("Invalid authorization")
}

// OptionFunc type
type OptionFunc func(*Middleware)

// SetTokenValidator option func
func SetTokenValidator(tokenValidator interfaces.TokenValidator) OptionFunc {
	return func(mw *Middleware) {
		mw.TokenValidator = tokenValidator
	}
}

// SetACLPermissionChecker option func
func SetACLPermissionChecker(aclPermissionChecker interfaces.ACLPermissionChecker) OptionFunc {
	return func(mw *Middleware) {
		mw.ACLPermissionChecker = aclPermissionChecker
	}
}

// SetBasicAuthValidator option func
func SetBasicAuthValidator(basicAuth interfaces.BasicAuthValidator) OptionFunc {
	return func(mw *Middleware) {
		mw.BasicAuthValidator = basicAuth
	}
}

type defaultMiddleware struct{}

func (defaultMiddleware) ValidateToken(ctx context.Context, token string) (*candishared.TokenClaim, error) {
	return &candishared.TokenClaim{}, nil
}
func (defaultMiddleware) CheckPermission(ctx context.Context, userID string, permissionCode string) (role string, err error) {
	return
}
func (defaultMiddleware) ValidateBasic(username, password string) error {
	if username != env.BaseEnv().BasicAuthUsername || password != env.BaseEnv().BasicAuthPassword {
		return errors.New("Invalid credentials")
	}
	return nil
}
