package interfaces

import (
	"context"

	"github.com/labstack/echo"
	"pkg.agungdwiprasetyo.com/candi/candishared"
)

// Middleware abstraction
type Middleware interface {
	Basic(context.Context, string) error
	Bearer(context.Context, string) (*candishared.TokenClaim, error)

	HTTPMiddleware
	GRPCMiddleware
	GraphQLMiddleware
}

// HTTPMiddleware interface, common middleware for http handler
type HTTPMiddleware interface {
	HTTPBasicAuth(showAlert bool) echo.MiddlewareFunc
	HTTPBearerAuth() echo.MiddlewareFunc
	HTTPMultipleAuth() echo.MiddlewareFunc
}

// GRPCMiddleware interface, common middleware for grpc handler
type GRPCMiddleware interface {
	GRPCBasicAuth(ctx context.Context)
	GRPCBearerAuth(ctx context.Context) *candishared.TokenClaim
}

// GraphQLMiddleware interface, common middleware for graphql handler, as directive in graphql schema
type GraphQLMiddleware interface {
	GraphQLBasicAuth(ctx context.Context) context.Context
	GraphQLBearerAuth(ctx context.Context) context.Context
}
