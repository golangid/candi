package middleware

import (
	"context"

	"agungdwiprasetyo.com/backend-microservices/config"
	"agungdwiprasetyo.com/backend-microservices/pkg/token"
	"github.com/labstack/echo"
	"google.golang.org/grpc"
)

// Middleware abstraction
type Middleware interface {
	BasicAuth(string) error
	ValidateBearer() echo.MiddlewareFunc
	GRPCAuth(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error)
	GRPCAuthStream(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error)
}

type mw struct {
	tokenUtil          token.Token
	username, password string
}

// NewMiddleware create new middleware instance
func NewMiddleware(cfg *config.Config) Middleware {
	return &mw{
		tokenUtil: token.NewJWT(cfg.PublicKey, cfg.PrivateKey),
		username:  config.BaseEnv().BasicAuthUsername,
		password:  config.BaseEnv().BasicAuthPassword,
	}
}
