package linechatbot

import (
	"context"

	"agungdwiprasetyo.com/backend-microservices/config"
	"agungdwiprasetyo.com/backend-microservices/config/broker"
	"agungdwiprasetyo.com/backend-microservices/config/database"
	"agungdwiprasetyo.com/backend-microservices/internal/line-chatbot/modules/chatbot"
	"agungdwiprasetyo.com/backend-microservices/internal/line-chatbot/modules/event"
	"agungdwiprasetyo.com/backend-microservices/pkg/codebase/factory"
	"agungdwiprasetyo.com/backend-microservices/pkg/codebase/factory/constant"
	"agungdwiprasetyo.com/backend-microservices/pkg/codebase/factory/dependency"
	"agungdwiprasetyo.com/backend-microservices/pkg/codebase/interfaces"
	"agungdwiprasetyo.com/backend-microservices/pkg/middleware"
	authsdk "agungdwiprasetyo.com/backend-microservices/pkg/sdk/auth-service"
)

// Service model
type Service struct {
	deps    dependency.Dependency
	modules []factory.ModuleFactory
	name    constant.Service
}

// NewService in this service
func NewService(serviceName string, cfg *config.Config) factory.ServiceFactory {
	var depsOptions = []dependency.Option{
		dependency.SetMiddleware(middleware.NewMiddleware(authsdk.NewAuthServiceGRPC())),
	}

	cfg.Load(
		func(ctx context.Context) interfaces.Closer {
			d := database.InitMongoDB(ctx)
			depsOptions = append(depsOptions, dependency.SetMongoDatabase(d))
			return d
		},
		func(context.Context) interfaces.Closer {
			d := broker.InitKafkaBroker(config.BaseEnv().Kafka.ClientID)
			depsOptions = append(depsOptions, dependency.SetBroker(d))
			return d
		},
	)

	// inject all service dependencies
	deps := dependency.InitDependency(depsOptions...)

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
