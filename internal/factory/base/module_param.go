package base

import (
	"github.com/agungdwiprasetyo/backend-microservices/config"
	"github.com/agungdwiprasetyo/backend-microservices/pkg/middleware"
)

// ModuleParam base
type ModuleParam struct {
	Config     *config.Config
	Middleware middleware.Middleware
}
