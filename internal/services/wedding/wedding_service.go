package wedding

import (
	"github.com/agungdwiprasetyo/backend-microservices/internal/factory"
	"github.com/agungdwiprasetyo/backend-microservices/internal/factory/base"
	"github.com/agungdwiprasetyo/backend-microservices/internal/factory/constant"
	"github.com/agungdwiprasetyo/backend-microservices/internal/services/wedding/modules/invitation"
)

const (
	// Wedding service name
	Wedding constant.Service = "wedding"
)

// Service model
type Service struct {
	modules []factory.ModuleFactory
}

// NewService in this service
func NewService(params *base.ModuleParam) factory.ServiceFactory {
	service := &Service{
		modules: []factory.ModuleFactory{
			invitation.NewModule(params),
		},
	}

	return service
}

// Modules method
func (s *Service) Modules() []factory.ModuleFactory {
	return s.modules
}

// Name method
func (s *Service) Name() constant.Service {
	return Wedding
}
