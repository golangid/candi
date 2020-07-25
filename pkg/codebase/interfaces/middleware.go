package interfaces

import (
	"context"

	"agungdwiprasetyo.com/backend-microservices/pkg/shared"
	"github.com/labstack/echo"
	"google.golang.org/grpc"
)

// Middleware abstraction
type Middleware interface {
	Basic(context.Context, string) error
	Bearer(context.Context, string) (*shared.TokenClaim, error)

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
	GRPCBasicAuth(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error)
	GRPCBasicAuthStream(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error)
}

// GraphQLMiddleware interface, common middleware for graphql handler, as directive in graphql schema
type GraphQLMiddleware interface {
	GraphQLBasicAuth(ctx context.Context)
	GraphQLBearerAuth(ctx context.Context) *shared.TokenClaim
}
