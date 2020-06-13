package main

const serviceMainTemplate = `package {{clean $.ServiceName}}

import (
	"{{.PackageName}}/internal/factory"
	"{{.PackageName}}/internal/factory/base"
	"{{.PackageName}}/internal/factory/constant"
{{- range $module := .Modules}}
	"{{$.PackageName}}/internal/services/{{$.ServiceName}}/modules/{{$module}}"
{{- end }}
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
