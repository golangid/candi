package main

const serviceMainTemplate = `// {{.Header}} DO NOT EDIT.

package {{clean $.ServiceName}}

import (
	"{{.PackageName}}/codebase/factory"
	"{{.PackageName}}/codebase/factory/dependency"
	"{{.PackageName}}/codebase/factory/types"
	"{{.PackageName}}/config"

	"{{$.GoModName}}/configs"
{{- range $module := .Modules}}
	"{{$.GoModName}}/internal/modules/{{$module.ModuleName}}"
{{- end }}
	"{{$.GoModName}}/pkg/shared/usecase"
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
	usecase.SetSharedUsecase(deps)

	modules := []factory.ModuleFactory{
	{{- range $module := .Modules}}
		{{clean $module.ModuleName}}.NewModule(deps),
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
