package factory

import (
	"pkg.agungdwiprasetyo.com/gendon/codebase/factory/types"
	"pkg.agungdwiprasetyo.com/gendon/codebase/interfaces"
)

// ModuleFactory factory
type ModuleFactory interface {
	RestHandler() interfaces.EchoRestHandler
	GRPCHandler() interfaces.GRPCHandler
	GraphQLHandler() interfaces.GraphQLHandler
	WorkerHandler(workerType types.Worker) interfaces.WorkerHandler
	Name() types.Module
}
