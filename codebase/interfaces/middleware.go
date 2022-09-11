package interfaces

import (
	"context"
	"net/http"

	"github.com/golangid/candi/candishared"
	"github.com/golangid/candi/codebase/factory/types"
	gqltypes "github.com/golangid/graphql-go/types"
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

	// HTTPPermissionACL method.
	// This middleware required TokenValidator (HTTPBearerAuth middleware must executed before) for extract userID
	// from token (from `Subject` field in token claim payload)
	HTTPPermissionACL(permissionCode string) func(http.Handler) http.Handler
}

// GRPCMiddleware interface, common middleware for grpc handler
type GRPCMiddleware interface {
	GRPCBasicAuth(ctx context.Context) context.Context
	GRPCBearerAuth(ctx context.Context) context.Context

	// GRPCPermissionACL method.
	// This middleware required TokenValidator (GRPCBearerAuth middleware must executed before) for extract userID
	// from token (from `Subject` field in token claim payload)
	GRPCPermissionACL(permissionCode string) types.MiddlewareFunc
}

// GraphQLMiddleware interface, common middleware for graphql handler, as directive in graphql schema
type GraphQLMiddleware interface {
	GraphQLAuth(ctx context.Context, directive *gqltypes.Directive, input interface{}) (context.Context, error)

	// GraphQLPermissionACL method.
	// This middleware required TokenValidator (GraphQLAuth middleware with BEARER must executed before) for extract userID
	// from token (from `Subject` field in token claim payload)
	GraphQLPermissionACL(ctx context.Context, directive *gqltypes.Directive, input interface{}) (context.Context, error)
}

// TokenValidator abstract interface for jwt validator
type TokenValidator interface {
	ValidateToken(ctx context.Context, token string) (*candishared.TokenClaim, error)
}

// ACLPermissionChecker abstraction for check acl permission with given permission code
type ACLPermissionChecker interface {
	CheckPermission(ctx context.Context, userID string, permissionCode string) (role string, err error)
}
