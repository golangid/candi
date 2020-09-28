package factory

import (
	"pkg.agungdwiprasetyo.com/candi/codebase/factory/dependency"
	"pkg.agungdwiprasetyo.com/candi/codebase/factory/types"
)

// ServiceFactory factory
type ServiceFactory interface {
	GetDependency() dependency.Dependency
	GetModules() []ModuleFactory
	Name() types.Service
}
