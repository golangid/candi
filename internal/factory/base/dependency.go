package base

import (
	"agungdwiprasetyo.com/backend-microservices/config"
	"agungdwiprasetyo.com/backend-microservices/pkg/middleware"
)

// Dependency base
type Dependency struct {
	Config     *config.Config
	Middleware middleware.Middleware
}

// InitDependency constructor
func InitDependency(cfg *config.Config) *Dependency {
	return &Dependency{
		Config:     cfg,
		Middleware: middleware.NewMiddleware(cfg),
	}
}
