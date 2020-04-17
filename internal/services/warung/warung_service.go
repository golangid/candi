package warung

import (
	"agungdwiprasetyo.com/backend-microservices/config"
	"agungdwiprasetyo.com/backend-microservices/internal/factory"
	"agungdwiprasetyo.com/backend-microservices/internal/factory/base"
	"agungdwiprasetyo.com/backend-microservices/internal/factory/constant"
	"agungdwiprasetyo.com/backend-microservices/internal/services/warung/modules/product"
	"agungdwiprasetyo.com/backend-microservices/internal/services/warung/modules/user"
)

// Service model
type Service struct {
	cfg  *config.Config
	name constant.Service
}

// NewService in this service
func NewService(cfg *config.Config, serviceName string) factory.ServiceFactory {
	service := &Service{
		cfg:  cfg,
		name: constant.Service(serviceName),
	}
	return service
}

// GetConfig method
func (s *Service) GetConfig() *config.Config {
	return s.cfg
}

// Modules method
func (s *Service) Modules(params *base.ModuleParam) []factory.ModuleFactory {
	return []factory.ModuleFactory{
		product.NewModule(params),
		user.NewModule(params),
	}
}

// Name method
func (s *Service) Name() constant.Service {
	return s.name
}
