package authservice

import (
	"context"

	"agungdwiprasetyo.com/backend-microservices/config"
	"agungdwiprasetyo.com/backend-microservices/config/broker"
	"agungdwiprasetyo.com/backend-microservices/config/database"
	"agungdwiprasetyo.com/backend-microservices/config/key"
	"agungdwiprasetyo.com/backend-microservices/internal/auth-service/modules/token"
	"agungdwiprasetyo.com/backend-microservices/pkg/codebase/factory"
	"agungdwiprasetyo.com/backend-microservices/pkg/codebase/factory/dependency"
	"agungdwiprasetyo.com/backend-microservices/pkg/codebase/factory/types"
	"agungdwiprasetyo.com/backend-microservices/pkg/codebase/interfaces"
	"agungdwiprasetyo.com/backend-microservices/pkg/middleware"
	"agungdwiprasetyo.com/backend-microservices/pkg/validator"
)

// Service model
type Service struct {
	deps    dependency.Dependency
	modules []factory.ModuleFactory
	name    types.Service
}

// NewService in this service
func NewService(serviceName string, cfg *config.Config) factory.ServiceFactory {
	// See all option in dependency package
	var deps dependency.Dependency

	cfg.LoadFunc(func(ctx context.Context) []interfaces.Closer {
		kafkaDeps := broker.InitKafkaBroker(config.BaseEnv().Kafka.Brokers, config.BaseEnv().Kafka.ClientID)
		redisDeps := database.InitRedis()
		rsaKey := key.LoadRSAKey()

		// inject all service dependencies
		deps = dependency.InitDependency(
			dependency.SetMiddleware(middleware.NewMiddleware(nil)),
			dependency.SetValidator(validator.NewValidator()),
			dependency.SetBroker(kafkaDeps),
			dependency.SetRedisPool(redisDeps),
			dependency.SetKey(rsaKey),
			// ... add more dependencies
		)
		return []interfaces.Closer{kafkaDeps, redisDeps} // throw back to config for close connection when application shutdown
	})

	modules := []factory.ModuleFactory{
		token.NewModule(deps),
	}

	return &Service{
		deps:    deps,
		modules: modules,
		name:    types.Service(serviceName),
	}
}

// GetDependency method
func (s *Service) GetDependency() dependency.Dependency {
	return s.deps
}

// GetModules method
func (s *Service) GetModules() []factory.ModuleFactory {
	return s.modules
}

// Name method
func (s *Service) Name() types.Service {
	return s.name
}
