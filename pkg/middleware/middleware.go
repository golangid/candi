package middleware

import (
	"context"

	"agungdwiprasetyo.com/backend-microservices/config"
	"agungdwiprasetyo.com/backend-microservices/pkg/shared"
	"agungdwiprasetyo.com/backend-microservices/pkg/token"
	"github.com/labstack/echo"
	"google.golang.org/grpc"
)

// Middleware abstraction
type Middleware interface {
	Basic(context.Context, string) error
	Bearer(context.Context, string) (*token.Claim, error)

	HTTPMiddleware
	GRPCMiddleware
	GraphQLMiddleware
}

// HTTPMiddleware interface, common middleware for http handler
type HTTPMiddleware interface {
	HTTPBasicAuth(showAlert bool) echo.MiddlewareFunc
	HTTPBearerAuth() echo.MiddlewareFunc
}

// GRPCMiddleware interface, common middleware for grpc handler
type GRPCMiddleware interface {
	GRPCBasicAuth(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error)
	GRPCBasicAuthStream(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error)
}

// GraphQLMiddleware interface, common middleware for graphql handler, as directive in graphql schema
type GraphQLMiddleware interface {
	GraphQLBasicAuth(ctx context.Context)
	GraphQLBearerAuth(ctx context.Context) *token.Claim
}

type mw struct {
	tokenValidator interface {
		Validate(ctx context.Context, token string) <-chan shared.Result
	}
	username, password string
}

// NewMiddleware create new middleware instance
func NewMiddleware(cfg *config.Config) Middleware {
	return &mw{
		tokenValidator: token.NewJWT(cfg.PublicKey, cfg.PrivateKey),
		username:       config.BaseEnv().BasicAuthUsername,
		password:       config.BaseEnv().BasicAuthPassword,
	}
}
