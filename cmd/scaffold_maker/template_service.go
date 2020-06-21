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
)

// Service model
type Service struct {
	deps    dependency.Dependency
	modules []factory.ModuleFactory
	name    constant.Service
}

// NewService in this service
func NewService(serviceName string, cfg *config.Config) factory.ServiceFactory {
	// See all optionn in dependency package
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
		// ... add some dependencies
	)

	// inject all service dependencies
	deps := dependency.InitDependency(depsOptions...)

	modules := []factory.ModuleFactory{
	{{- range $module := .Modules}}
		{{clean $module}}.NewModule(deps),
	{{- end }}
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

`
