package factory

import (
	"github.com/golangid/candi/codebase/factory/dependency"
	"github.com/golangid/candi/codebase/factory/types"
	"github.com/golangid/candi/config"
)

// ServiceFactory factory
type ServiceFactory interface {
	GetConfig() *config.Config
	GetDependency() dependency.Dependency
	GetApplications() []AppServerFactory
	GetModules() []ModuleFactory
	Name() types.Service
}
