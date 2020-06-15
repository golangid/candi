package main

const serviceMainTemplate = `package {{clean $.ServiceName}}

import (
{{- range $module := .Modules}}
	"{{$.PackageName}}/internal/{{$.ServiceName}}/modules/{{$module}}"
{{- end }}
	"{{.PackageName}}/pkg/codebase/factory"
	"{{.PackageName}}/pkg/codebase/factory/base"
	"{{.PackageName}}/pkg/codebase/factory/constant"
)

// Service model
type Service struct {
	dependency *base.Dependency
	modules    []factory.ModuleFactory
	name       constant.Service
}

// NewService in this service
func NewService(serviceName string, dependency *base.Dependency) factory.ServiceFactory {
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
func (s *Service) GetDependency() *base.Dependency {
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
