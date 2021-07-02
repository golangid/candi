package factory

import (
	"pkg.agungdp.dev/candi/codebase/factory/types"
	"pkg.agungdp.dev/candi/codebase/interfaces"
)

// ModuleFactory factory
type ModuleFactory interface {
	// Basic server type (current using echo rest framework)
	RESTHandler() interfaces.RESTHandler
	GRPCHandler() interfaces.GRPCHandler
	GraphQLHandler() interfaces.GraphQLHandler
	WorkerHandler(workerType types.Worker) interfaces.WorkerHandler
	Name() types.Module

	// Additional server type (another rest framework, p2p, and many more)
	ServerHandler(serverType types.Server) interfaces.ServerHandler
}
