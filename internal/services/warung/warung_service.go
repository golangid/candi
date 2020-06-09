package warung

import (
	"agungdwiprasetyo.com/backend-microservices/internal/factory"
	"agungdwiprasetyo.com/backend-microservices/internal/factory/base"
	"agungdwiprasetyo.com/backend-microservices/internal/factory/constant"
	"agungdwiprasetyo.com/backend-microservices/internal/services/warung/modules/product"
	"agungdwiprasetyo.com/backend-microservices/internal/services/warung/modules/user"
)

// Service model
type Service struct {
	dependency *base.Dependency
	modules    []factory.ModuleFactory
	name       constant.Service
}

// NewService in this service
func NewService(serviceName string, dependency *base.Dependency) factory.ServiceFactory {
	modules := []factory.ModuleFactory{
		product.NewModule(dependency),
		user.NewModule(dependency),
	}

	return &Service{
		dependency: dependency,
		modules:    modules,
		name:       constant.Service(serviceName),
	}
}

// GetDependency method
func (s *Service) GetDependency() *base.Dependency {
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
