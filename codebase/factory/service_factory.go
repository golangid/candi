package factory

import (
	"pkg.agungdp.dev/candi/codebase/factory/dependency"
	"pkg.agungdp.dev/candi/codebase/factory/types"
	"pkg.agungdp.dev/candi/config"
)

// ServiceFactory factory
type ServiceFactory interface {
	GetConfig() *config.Config
	GetDependency() dependency.Dependency
	GetApplications() []AppServerFactory
	GetModules() []ModuleFactory
	Name() types.Service
}
