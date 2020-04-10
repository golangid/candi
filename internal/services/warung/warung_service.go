package warung

import (
	"agungdwiprasetyo.com/backend-microservices/config"
	"agungdwiprasetyo.com/backend-microservices/internal/factory"
	"agungdwiprasetyo.com/backend-microservices/internal/factory/base"
	"agungdwiprasetyo.com/backend-microservices/internal/factory/constant"
	"agungdwiprasetyo.com/backend-microservices/internal/services/warung/modules/product"
	"agungdwiprasetyo.com/backend-microservices/internal/services/warung/modules/user"
)

const (
	// Warung service name
	Warung constant.Service = "warung"
)

// Service model
type Service struct {
	cfg *config.Config
}

// NewService in this service
func NewService(cfg *config.Config) factory.ServiceFactory {
	service := &Service{
		cfg: cfg,
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
	return Warung
}
