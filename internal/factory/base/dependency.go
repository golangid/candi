package base

import (
	"agungdwiprasetyo.com/backend-microservices/config"
	"agungdwiprasetyo.com/backend-microservices/pkg/middleware"
	"agungdwiprasetyo.com/backend-microservices/pkg/publisher"
)

// Dependency base
type Dependency struct {
	Config     *config.Config
	Middleware middleware.Middleware
	Publisher  publisher.Publisher
}

// InitDependency constructor
func InitDependency(cfg *config.Config) *Dependency {
	return &Dependency{
		Config:     cfg,
		Middleware: middleware.NewMiddleware(cfg),
		Publisher:  publisher.NewKafkaPublisher(cfg.KafkaConfig),
	}
}
