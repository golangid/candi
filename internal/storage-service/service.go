package service

import (
	"context"
	"os"

	"agungdwiprasetyo.com/backend-microservices/config"
	"agungdwiprasetyo.com/backend-microservices/config/database"
	"agungdwiprasetyo.com/backend-microservices/internal/storage-service/modules/storage"
	"agungdwiprasetyo.com/backend-microservices/pkg/codebase/factory"
	"agungdwiprasetyo.com/backend-microservices/pkg/codebase/factory/constant"
	"agungdwiprasetyo.com/backend-microservices/pkg/codebase/factory/dependency"
	"agungdwiprasetyo.com/backend-microservices/pkg/codebase/interfaces"
	"agungdwiprasetyo.com/backend-microservices/pkg/middleware"
	authsdk "agungdwiprasetyo.com/backend-microservices/pkg/sdk/auth-service"
	"agungdwiprasetyo.com/backend-microservices/pkg/validator"
)

// Service model
type Service struct {
	deps    dependency.Dependency
	modules []factory.ModuleFactory
	name    constant.Service
}

// NewService in this service
func NewService(serviceName string, cfg *config.Config) factory.ServiceFactory {
	// See all option in dependency package
	var deps dependency.Dependency

	cfg.LoadFunc(func(ctx context.Context) []interfaces.Closer {
		redisDeps := database.InitRedis()

		authService := authsdk.NewAuthServiceGRPC(os.Getenv("AUTH_SERVICE_GRPC_HOST"), os.Getenv("AUTH_SERVICE_GRPC_KEY"))

		// inject all service dependencies
		deps = dependency.InitDependency(
			dependency.SetMiddleware(middleware.NewMiddleware(authService)),
			dependency.SetValidator(validator.NewJSONSchemaValidator(serviceName)),
			dependency.SetRedisPool(redisDeps),
			// ... add more dependencies
		)
		return []interfaces.Closer{redisDeps} // throw back to config for close connection when application shutdown
	})

	modules := []factory.ModuleFactory{
		storage.NewModule(deps),
	}

	return &Service{
		deps:    deps,
		modules: modules,
		name:    constant.Service(serviceName),
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
func (s *Service) Name() constant.Service {
	return s.name
}
