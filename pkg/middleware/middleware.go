package middleware

import (
	"agungdwiprasetyo.com/backend-microservices/config"
	"agungdwiprasetyo.com/backend-microservices/pkg/codebase/interfaces"
)

type mw struct {
	tokenValidator     interfaces.TokenValidator
	username, password string
}

// NewMiddleware create new middleware instance
func NewMiddleware(tokenValidator interfaces.TokenValidator) interfaces.Middleware {
	return &mw{
		tokenValidator: tokenValidator,
		username:       config.BaseEnv().BasicAuthUsername,
		password:       config.BaseEnv().BasicAuthPassword,
	}
}
