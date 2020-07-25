package userservice

import (
	"context"
	"os"

	"agungdwiprasetyo.com/backend-microservices/config"
	"agungdwiprasetyo.com/backend-microservices/config/broker"
	"agungdwiprasetyo.com/backend-microservices/config/database"
	"agungdwiprasetyo.com/backend-microservices/internal/user-service/modules/auth"
	"agungdwiprasetyo.com/backend-microservices/internal/user-service/modules/customer"
	"agungdwiprasetyo.com/backend-microservices/internal/user-service/modules/member"
	"agungdwiprasetyo.com/backend-microservices/pkg/codebase/factory"
	"agungdwiprasetyo.com/backend-microservices/pkg/codebase/factory/dependency"
	"agungdwiprasetyo.com/backend-microservices/pkg/codebase/factory/types"
	"agungdwiprasetyo.com/backend-microservices/pkg/codebase/interfaces"
	"agungdwiprasetyo.com/backend-microservices/pkg/middleware"
	authsdk "agungdwiprasetyo.com/backend-microservices/pkg/sdk/auth-service"
	"agungdwiprasetyo.com/backend-microservices/pkg/validator"
)

// Service model
type Service struct {
	deps    dependency.Dependency
	modules []factory.ModuleFactory
	name    types.Service
}

// NewService starting service
func NewService(serviceName string, cfg *config.Config) factory.ServiceFactory {
	// See all option in dependency package
	var deps dependency.Dependency

	cfg.LoadFunc(func(ctx context.Context) []interfaces.Closer {
		kafkaDeps := broker.InitKafkaBroker(config.BaseEnv().Kafka.Brokers, config.BaseEnv().Kafka.ClientID)
		redisDeps := database.InitRedis()
		mongoDeps := database.InitMongoDB(ctx)
		sqlDeps := database.InitSQLDatabase()

		authService := authsdk.NewAuthServiceGRPC(os.Getenv("AUTH_SERVICE_GRPC_HOST"), os.Getenv("AUTH_SERVICE_GRPC_KEY"))

		// inject all service dependencies
		deps = dependency.InitDependency(
			dependency.SetMiddleware(middleware.NewMiddleware(authService)),
			dependency.SetValidator(validator.NewValidator()),
			dependency.SetBroker(kafkaDeps),
			dependency.SetRedisPool(redisDeps),
			dependency.SetMongoDatabase(mongoDeps),
			dependency.SetSQLDatabase(sqlDeps),
			// ... add more dependencies
		)
		return []interfaces.Closer{kafkaDeps, redisDeps, mongoDeps, sqlDeps} // throw back to config for close connection when application shutdown
	})

	modules := []factory.ModuleFactory{
		member.NewModule(deps),
		customer.NewModule(deps),
		auth.NewModule(deps),
	}

	return &Service{
		deps:    deps,
		modules: modules,
		name:    types.Service(serviceName),
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
func (s *Service) Name() types.Service {
	return s.name
}
