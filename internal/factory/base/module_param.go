package base

import (
	"agungdwiprasetyo.com/backend-microservices/config"
	"agungdwiprasetyo.com/backend-microservices/pkg/middleware"
)

// ModuleParam base
type ModuleParam struct {
	Config     *config.Config
	Middleware middleware.Middleware
}
