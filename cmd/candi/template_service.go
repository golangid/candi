package main

const serviceMainTemplate = `package {{clean $.ServiceName}}

import (
	"context"

	"pkg.agungdwiprasetyo.com/candi/codebase/factory"
	"pkg.agungdwiprasetyo.com/candi/codebase/factory/dependency"
	"pkg.agungdwiprasetyo.com/candi/codebase/factory/types"
	"pkg.agungdwiprasetyo.com/candi/codebase/interfaces"
	"pkg.agungdwiprasetyo.com/candi/config"
	"pkg.agungdwiprasetyo.com/candi/config/broker"
	"pkg.agungdwiprasetyo.com/candi/config/database"
	"pkg.agungdwiprasetyo.com/candi/middleware"
	"pkg.agungdwiprasetyo.com/candi/validator"


{{- range $module := .Modules}}
	"{{$.ServiceName}}/internal/modules/{{$module}}"
{{- end }}
)

// Service model
type Service struct {
	deps    dependency.Dependency
	modules []factory.ModuleFactory
	name    types.Service
}

// NewService in this service
func NewService(serviceName string, cfg *config.Config) factory.ServiceFactory {
	// See all option in dependency package
	var deps dependency.Dependency

	cfg.LoadFunc(func(ctx context.Context) []interfaces.Closer {
		kafkaDeps := broker.InitKafkaBroker(config.BaseEnv().Kafka.Brokers, config.BaseEnv().Kafka.ClientID)
		redisDeps := database.InitRedis()
		mongoDeps := database.InitMongoDB(ctx)

		// inject all service dependencies
		deps = dependency.InitDependency(
			dependency.SetMiddleware(middleware.NewMiddleware(nil)),
			dependency.SetValidator(validator.NewValidator()),
			dependency.SetBroker(kafkaDeps),
			dependency.SetRedisPool(redisDeps),
			dependency.SetMongoDatabase(mongoDeps),
			// ... add more dependencies
		)
		return []interfaces.Closer{kafkaDeps, redisDeps, mongoDeps} // throw back to config for close connection when application shutdown
	})

	modules := []factory.ModuleFactory{
	{{- range $module := .Modules}}
		{{clean $module}}.NewModule(deps),
	{{- end }}
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

`
