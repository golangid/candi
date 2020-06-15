package wedding

import (
	"agungdwiprasetyo.com/backend-microservices/internal/services/wedding/modules/event"
	"agungdwiprasetyo.com/backend-microservices/internal/services/wedding/modules/invitation"
	"agungdwiprasetyo.com/backend-microservices/pkg/codebase/factory"
	"agungdwiprasetyo.com/backend-microservices/pkg/codebase/factory/base"
	"agungdwiprasetyo.com/backend-microservices/pkg/codebase/factory/constant"
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
		invitation.NewModule(dependency),
		event.NewModule(dependency),
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
