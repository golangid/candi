package factory

import (
	"pkg.agungdwiprasetyo.com/gendon/codebase/factory/dependency"
	"pkg.agungdwiprasetyo.com/gendon/codebase/factory/types"
)

// ServiceFactory factory
type ServiceFactory interface {
	GetDependency() dependency.Dependency
	GetModules() []ModuleFactory
	Name() types.Service
}
