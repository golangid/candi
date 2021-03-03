package interfaces

import (
	"context"
	"net/http"

	"pkg.agungdp.dev/candi/candishared"
	"pkg.agungdp.dev/candi/codebase/factory/types"
)

// Middleware abstraction
type Middleware interface {
	Basic(ctx context.Context, authKey string) error
	Bearer(ctx context.Context, token string) (*candishared.TokenClaim, error)

	HTTPMiddleware
	GRPCMiddleware
	GraphQLMiddleware
}

// HTTPMiddleware interface, common middleware for http handler
type HTTPMiddleware interface {
	HTTPBasicAuth(next http.Handler) http.Handler
	HTTPBearerAuth(next http.Handler) http.Handler
	HTTPMultipleAuth(next http.Handler) http.Handler
	HTTPPermissionACL(permissionCode string) func(http.Handler) http.Handler
}

// GRPCMiddleware interface, common middleware for grpc handler
type GRPCMiddleware interface {
	GRPCBasicAuth(ctx context.Context) context.Context
	GRPCBearerAuth(ctx context.Context) context.Context
	GRPCPermissionACL(permissionCode string) types.MiddlewareFunc
}

// GraphQLMiddleware interface, common middleware for graphql handler, as directive in graphql schema
type GraphQLMiddleware interface {
	GraphQLBasicAuth(ctx context.Context) context.Context
	GraphQLBearerAuth(ctx context.Context) context.Context
	GraphQLPermissionACL(permissionCode string) types.MiddlewareFunc
}

// TokenValidator abstract interface for jwt validator
type TokenValidator interface {
	ValidateToken(ctx context.Context, token string) (*candishared.TokenClaim, error)
}

// ACLPermissionChecker abstraction for check acl permission with given permission code
type ACLPermissionChecker interface {
	CheckPermission(ctx context.Context, userID string, permissionCode string) (role string, err error)
}
