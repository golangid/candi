package factory

import (
	"agungdwiprasetyo.com/backend-microservices/pkg/codebase/factory/base"
	"agungdwiprasetyo.com/backend-microservices/pkg/codebase/factory/constant"
)

// ServiceFactory factory
type ServiceFactory interface {
	GetDependency() base.Dependency
	GetModules() []ModuleFactory
	Name() constant.Service
}
