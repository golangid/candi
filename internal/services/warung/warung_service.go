package warung

import (
	"github.com/agungdwiprasetyo/backend-microservices/internal/factory"
	"github.com/agungdwiprasetyo/backend-microservices/internal/factory/base"
	"github.com/agungdwiprasetyo/backend-microservices/internal/factory/constant"
	"github.com/agungdwiprasetyo/backend-microservices/internal/services/warung/modules/product"
	"github.com/agungdwiprasetyo/backend-microservices/internal/services/warung/modules/user"
)

const (
	// Warung service name
	Warung constant.Service = "warung"
)

// Service model
type Service struct {
	modules []factory.ModuleFactory
}

// NewService in this service
func NewService(params *base.ModuleParam) factory.ServiceFactory {
	service := &Service{
		modules: []factory.ModuleFactory{
			product.NewModule(params),
			user.NewModule(params),
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
	return Warung
}
