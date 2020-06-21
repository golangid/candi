package factory

import (
	"agungdwiprasetyo.com/backend-microservices/pkg/codebase/factory/constant"
	"agungdwiprasetyo.com/backend-microservices/pkg/codebase/factory/dependency"
)

// ServiceFactory factory
type ServiceFactory interface {
	GetDependency() dependency.Dependency
	GetModules() []ModuleFactory
	Name() constant.Service
}
