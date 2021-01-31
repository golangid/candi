package factory

import (
	"pkg.agungdp.dev/candi/codebase/factory/dependency"
	"pkg.agungdp.dev/candi/codebase/factory/types"
)

// ServiceFactory factory
type ServiceFactory interface {
	GetDependency() dependency.Dependency
	GetModules() []ModuleFactory
	Name() types.Service
}
