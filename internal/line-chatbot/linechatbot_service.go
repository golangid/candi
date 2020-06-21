package linechatbot

import (
	"agungdwiprasetyo.com/backend-microservices/config"
	"agungdwiprasetyo.com/backend-microservices/internal/line-chatbot/modules/chatbot"
	"agungdwiprasetyo.com/backend-microservices/internal/line-chatbot/modules/event"
	"agungdwiprasetyo.com/backend-microservices/pkg/codebase/factory"
	"agungdwiprasetyo.com/backend-microservices/pkg/codebase/factory/constant"
	"agungdwiprasetyo.com/backend-microservices/pkg/codebase/factory/dependency"
	"agungdwiprasetyo.com/backend-microservices/pkg/middleware"
	"agungdwiprasetyo.com/backend-microservices/pkg/publisher"
)

// Service model
type Service struct {
	deps    dependency.Dependency
	modules []factory.ModuleFactory
	name    constant.Service
}

// NewService in this service
func NewService(serviceName string, cfg *config.Config) factory.ServiceFactory {
	// init all service dependencies
	deps := dependency.InitDependency(
		dependency.SetMiddleware(middleware.NewMiddleware(nil)),
		dependency.SetMongoDatabase(cfg.MongoDB),
		dependency.SetBroker(cfg.KafkaConfig, publisher.NewKafkaPublisher(cfg.KafkaConfig)),
	)

	modules := []factory.ModuleFactory{
		chatbot.NewModule(deps),
		event.NewModule(deps),
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
