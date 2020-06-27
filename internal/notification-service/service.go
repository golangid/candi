package notificationservice

import (
	"context"

	"agungdwiprasetyo.com/backend-microservices/config"
	"agungdwiprasetyo.com/backend-microservices/config/broker"
	"agungdwiprasetyo.com/backend-microservices/config/database"
	pushnotif "agungdwiprasetyo.com/backend-microservices/internal/notification-service/modules/push-notif"
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
	// See all optionn in dependency package
	var depsOptions = []dependency.Option{
		dependency.SetMiddleware(middleware.NewMiddleware(authsdk.NewAuthServiceGRPC())),
		dependency.SetValidator(validator.NewJSONSchemaValidator(serviceName)),
	}

	cfg.Load(
		func(context.Context) interfaces.Closer {
			d := broker.InitKafkaBroker(config.BaseEnv().Kafka.ClientID)
			depsOptions = append(depsOptions, dependency.SetBroker(d))
			return d
		},
		func(context.Context) interfaces.Closer {
			d := database.InitRedis()
			depsOptions = append(depsOptions, dependency.SetRedisPool(d))
			return d
		},
		// ... add some dependencies
	)

	// inject all service dependencies
	deps := dependency.InitDependency(depsOptions...)

	modules := []factory.ModuleFactory{
		pushnotif.NewModule(deps),
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
