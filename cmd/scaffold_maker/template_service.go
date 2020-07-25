package main

const serviceMainTemplate = `package {{clean $.ServiceName}}

import (
	"context"

	"{{$.PackageName}}/config"
	"{{$.PackageName}}/config/broker"
	"{{$.PackageName}}/config/database"
{{- range $module := .Modules}}
	"{{$.PackageName}}/internal/{{$.ServiceName}}/modules/{{$module}}"
{{- end }}
	"{{$.PackageName}}/pkg/codebase/factory"
	"{{$.PackageName}}/pkg/codebase/factory/constant"
	"{{$.PackageName}}/pkg/codebase/factory/dependency"
	"{{$.PackageName}}/pkg/codebase/interfaces"
	"{{$.PackageName}}/pkg/middleware"
	authsdk "{{$.PackageName}}/pkg/sdk/auth-service"
	"{{$.PackageName}}/pkg/validator"
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
		kafkaDeps := broker.InitKafkaBroker(config.BaseEnv().Kafka.ClientID)
		redisDeps := database.InitRedis()
		mongoDeps := database.InitMongoDB(ctx)

		// inject all service dependencies
		deps = dependency.InitDependency(
			dependency.SetMiddleware(middleware.NewMiddleware(authsdk.NewAuthServiceGRPC())),
			dependency.SetValidator(validator.NewJSONSchemaValidator(serviceName)),
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
