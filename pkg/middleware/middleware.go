package middleware

import (
	"context"

	"github.com/agungdwiprasetyo/backend-microservices/config"
	"github.com/agungdwiprasetyo/backend-microservices/pkg/token"
	"github.com/labstack/echo"
	"google.golang.org/grpc"
)

// Middleware abstraction
type Middleware interface {
	BasicAuth() echo.MiddlewareFunc
	ValidateBearer() echo.MiddlewareFunc
	GRPCAuth(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error)
	GRPCAuthStream(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error)
}

type mw struct {
	tokenUtil          token.Token
	username, password string
	grpcAuthKey        string
}

// NewMiddleware create new middleware instance
func NewMiddleware(cfg *config.Config) Middleware {
	return &mw{
		tokenUtil:   token.NewJWT(cfg.PublicKey, cfg.PrivateKey),
		username:    config.GlobalEnv.BasicAuthUsername,
		password:    config.GlobalEnv.BasicAuthPassword,
		grpcAuthKey: config.GlobalEnv.GRPCAuthKey,
	}
}
