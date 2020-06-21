package userservice

import (
	"agungdwiprasetyo.com/backend-microservices/config"
	"agungdwiprasetyo.com/backend-microservices/internal/user-service/modules/auth"
	"agungdwiprasetyo.com/backend-microservices/internal/user-service/modules/customer"
	"agungdwiprasetyo.com/backend-microservices/internal/user-service/modules/member"
	"agungdwiprasetyo.com/backend-microservices/pkg/codebase/factory"
	"agungdwiprasetyo.com/backend-microservices/pkg/codebase/factory/base"
	"agungdwiprasetyo.com/backend-microservices/pkg/codebase/factory/constant"
	"agungdwiprasetyo.com/backend-microservices/pkg/middleware"
	"agungdwiprasetyo.com/backend-microservices/pkg/publisher"
	authsdk "agungdwiprasetyo.com/backend-microservices/pkg/sdk/auth-service"
)

// Service model
type Service struct {
	dependency base.Dependency
	modules    []factory.ModuleFactory
	name       constant.Service
}

// NewService starting service
func NewService(serviceName string, cfg *config.Config) factory.ServiceFactory {
	// init all service dependencies
	deps := base.InitDependency(
		base.SetMiddleware(middleware.NewMiddleware(authsdk.NewAuthServiceGRPC())),
		base.SetMongoDatabase(cfg.MongoDB),
		base.SetBroker(cfg.KafkaConfig, publisher.NewKafkaPublisher(cfg.KafkaConfig)),
	)

	modules := []factory.ModuleFactory{
		member.NewModule(deps),
		customer.NewModule(deps),
		auth.NewModule(deps),
	}

	return &Service{
		dependency: deps,
		modules:    modules,
		name:       constant.Service(serviceName),
	}
}

// GetDependency method
func (s *Service) GetDependency() base.Dependency {
	return s.dependency
}

// GetModules method
func (s *Service) GetModules() []factory.ModuleFactory {
	return s.modules
}

// Name method
func (s *Service) Name() constant.Service {
	return s.name
}
