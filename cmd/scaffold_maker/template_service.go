package main

const serviceMainTemplate = `package {{clean $.ServiceName}}

import (
{{- range $module := .Modules}}
	"{{$.PackageName}}/internal/{{$.ServiceName}}/modules/{{$module}}"
{{- end }}
	"{{.PackageName}}/pkg/codebase/factory"
	"{{.PackageName}}/pkg/codebase/factory/dependency"
	"{{.PackageName}}/pkg/codebase/factory/constant"
)

// Service model
type Service struct {
	dependency *dependency.Dependency
	modules    []factory.ModuleFactory
	name       constant.Service
}

// NewService in this service
func NewService(serviceName string, dependency *dependency.Dependency) factory.ServiceFactory {
	modules := []factory.ModuleFactory{
	{{- range $module := .Modules}}
		{{clean $module}}.NewModule(dependency),
	{{- end }}
	}

	return &Service{
		dependency: dependency,
		modules:    modules,
		name:       constant.Service(serviceName),
	}
}

// GetDependency method
func (s *Service) GetDependency() *dependency.Dependency {
	return s.dependency
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
