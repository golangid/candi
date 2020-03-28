package middleware

import (
	"github.com/agungdwiprasetyo/backend-microservices/config"
)

// Middleware model
type Middleware struct {
	username, password string
	grpcAuthKey        string
}

// NewMiddleware create new middleware instance
func NewMiddleware(cfg *config.Config) *Middleware {
	return &Middleware{
		username:    config.GlobalEnv.BasicAuthUsername,
		password:    config.GlobalEnv.BasicAuthPassword,
		grpcAuthKey: config.GlobalEnv.GRPCAuthKey,
	}
}
