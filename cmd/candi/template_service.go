package main

const serviceMainTemplate = `package {{clean $.ServiceName}}

import (
	"pkg.agungdwiprasetyo.com/candi/codebase/factory"
	"pkg.agungdwiprasetyo.com/candi/codebase/factory/dependency"
	"pkg.agungdwiprasetyo.com/candi/codebase/factory/types"
	"pkg.agungdwiprasetyo.com/candi/config"

	"{{$.ServiceName}}/configs"
{{- range $module := .Modules}}
	"{{$.ServiceName}}/internal/modules/{{$module.Name}}"
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
	deps := configs.LoadConfigs(cfg)

	modules := []factory.ModuleFactory{
	{{- range $module := .Modules}}
		{{clean $module.Name}}.NewModule(deps),
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
