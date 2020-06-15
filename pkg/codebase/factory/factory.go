package factory

import (
	"agungdwiprasetyo.com/backend-microservices/pkg/codebase/factory/base"
	"agungdwiprasetyo.com/backend-microservices/pkg/codebase/factory/constant"
	"agungdwiprasetyo.com/backend-microservices/pkg/codebase/factory/interfaces"
)

// ServiceFactory factory
type ServiceFactory interface {
	GetDependency() *base.Dependency
	GetModules() []ModuleFactory
	Name() constant.Service
}

// ModuleFactory factory
type ModuleFactory interface {
	RestHandler() interfaces.EchoRestHandler
	GRPCHandler() interfaces.GRPCHandler
	GraphQLHandler() (name string, resolver interface{})
	WorkerHandler(workerType constant.Worker) interfaces.WorkerHandler
	Name() constant.Module
}
