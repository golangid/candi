package main

const serviceMainTemplate = `package {{clean $.ServiceName}}

import (
	"{{.PackageName}}/config"
	"{{.PackageName}}/internal/factory"
	"{{.PackageName}}/internal/factory/base"
	"{{.PackageName}}/internal/factory/constant"
{{- range $module := .Modules}}
	"{{$.PackageName}}/internal/services/{{$.ServiceName}}/modules/{{$module}}"
{{- end }}
)

// Service model
type Service struct {
	cfg  *config.Config
	name constant.Service
}

// NewService in this service
func NewService(cfg *config.Config, serviceName string) factory.ServiceFactory {
	service := &Service{
		cfg:  cfg,
		name: constant.Service(serviceName),
	}
	return service
}

// GetConfig method
func (s *Service) GetConfig() *config.Config {
	return s.cfg
}

// Modules method
func (s *Service) Modules(params *base.ModuleParam) []factory.ModuleFactory {
	return []factory.ModuleFactory{
	{{- range $module := .Modules}}
		{{clean $module}}.NewModule(params),
	{{- end }}
	}
}

// Name method
func (s *Service) Name() constant.Service {
	return s.name
}

`
