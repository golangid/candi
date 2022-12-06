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
	tokenValidator       interfaces.TokenValidator
	aclPermissionChecker interfaces.ACLPermissionChecker
	basicAuthValidator   interfaces.BasicAuthValidator

	extractUserIDFunc func(tokenClaim *candishared.TokenClaim) (userID string)
}

// NewMiddleware create new middleware instance (DEPRECATED, use NewMiddlewareWithOption)
func NewMiddleware(tokenValidator interfaces.TokenValidator, aclPermissionChecker interfaces.ACLPermissionChecker) *Middleware {
	mw := &Middleware{
		tokenValidator: tokenValidator, aclPermissionChecker: aclPermissionChecker,
		basicAuthValidator: &defaultMiddleware{},
		extractUserIDFunc: func(tokenClaim *candishared.TokenClaim) (userID string) {
			return tokenClaim.Subject
		},
	}

	return mw
}

// NewMiddlewareWithOption create new middleware instance with option
func NewMiddlewareWithOption(opts ...OptionFunc) *Middleware {
	defaultMw := &defaultMiddleware{}
	mw := &Middleware{
		tokenValidator: defaultMw, aclPermissionChecker: defaultMw, basicAuthValidator: defaultMw,
		extractUserIDFunc: func(tokenClaim *candishared.TokenClaim) (userID string) {
			return tokenClaim.Subject
		},
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
		mw.tokenValidator = tokenValidator
	}
}

// SetACLPermissionChecker option func
func SetACLPermissionChecker(aclPermissionChecker interfaces.ACLPermissionChecker) OptionFunc {
	return func(mw *Middleware) {
		mw.aclPermissionChecker = aclPermissionChecker
	}
}

// SetBasicAuthValidator option func
func SetBasicAuthValidator(basicAuth interfaces.BasicAuthValidator) OptionFunc {
	return func(mw *Middleware) {
		mw.basicAuthValidator = basicAuth
	}
}

// SetUserIDExtractor option func, custom extract user id from token claim for acl permission checker
func SetUserIDExtractor(extractor func(tokenClaim *candishared.TokenClaim) (userID string)) OptionFunc {
	return func(mw *Middleware) {
		mw.extractUserIDFunc = extractor
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
