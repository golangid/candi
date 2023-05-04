package middleware

import (
	"context"
	"errors"
	"time"

	"github.com/golangid/candi/candishared"
	"github.com/golangid/candi/codebase/interfaces"
	"github.com/golangid/candi/config/env"
)

// Middleware impl
type Middleware struct {
	tokenValidator       interfaces.TokenValidator
	aclPermissionChecker interfaces.ACLPermissionChecker
	basicAuthValidator   interfaces.BasicAuthValidator

	cache           interfaces.Cache
	defaultCacheAge time.Duration

	extractUserIDFunc func(tokenClaim *candishared.TokenClaim) (userID string)
}

// NewMiddleware create new middleware instance (DEPRECATED, use NewMiddlewareWithOption)
func NewMiddleware(tokenValidator interfaces.TokenValidator, aclPermissionChecker interfaces.ACLPermissionChecker) *Middleware {
	mw := &Middleware{
		tokenValidator: tokenValidator, aclPermissionChecker: aclPermissionChecker,
		basicAuthValidator: &defaultMiddleware{
			username: env.BaseEnv().BasicAuthUsername, password: env.BaseEnv().BasicAuthPassword,
		},
		extractUserIDFunc: func(tokenClaim *candishared.TokenClaim) (userID string) {
			return tokenClaim.Subject
		},
		defaultCacheAge: DefaultCacheAge,
	}

	return mw
}

// NewMiddlewareWithOption create new middleware instance with option
func NewMiddlewareWithOption(opts ...OptionFunc) *Middleware {
	defaultMw := &defaultMiddleware{
		username: env.BaseEnv().BasicAuthUsername, password: env.BaseEnv().BasicAuthPassword,
	}
	mw := &Middleware{
		tokenValidator: defaultMw, aclPermissionChecker: defaultMw, basicAuthValidator: defaultMw,
		extractUserIDFunc: func(tokenClaim *candishared.TokenClaim) (userID string) {
			return tokenClaim.Subject
		},
		defaultCacheAge: DefaultCacheAge,
	}
	for _, opt := range opts {
		opt(mw)
	}

	return mw
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

// SetCache option func
func SetCache(cache interfaces.Cache, defaultCacheAge time.Duration) OptionFunc {
	return func(mw *Middleware) {
		mw.cache = cache
		mw.defaultCacheAge = defaultCacheAge
	}
}

// SetUserIDExtractor option func, custom extract user id from token claim for acl permission checker
func SetUserIDExtractor(extractor func(tokenClaim *candishared.TokenClaim) (userID string)) OptionFunc {
	return func(mw *Middleware) {
		mw.extractUserIDFunc = extractor
	}
}

type defaultMiddleware struct {
	username, password string
}

func (defaultMiddleware) ValidateToken(ctx context.Context, token string) (*candishared.TokenClaim, error) {
	return &candishared.TokenClaim{}, nil
}
func (defaultMiddleware) CheckPermission(ctx context.Context, userID string, permissionCode string) (role string, err error) {
	return
}
func (d *defaultMiddleware) ValidateBasic(ctx context.Context, username, password string) error {
	if username != d.username || password != d.password {
		return errors.New("Invalid credentials")
	}
	return nil
}
