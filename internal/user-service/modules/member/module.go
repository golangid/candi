package member

import (
	"agungdwiprasetyo.com/backend-microservices/internal/user-service/modules/member/delivery/graphqlhandler"
	"agungdwiprasetyo.com/backend-microservices/internal/user-service/modules/member/delivery/grpchandler"
	"agungdwiprasetyo.com/backend-microservices/internal/user-service/modules/member/delivery/resthandler"
	"agungdwiprasetyo.com/backend-microservices/pkg/codebase/factory/dependency"
	"agungdwiprasetyo.com/backend-microservices/pkg/codebase/factory/types"
	"agungdwiprasetyo.com/backend-microservices/pkg/codebase/interfaces"
)

const (
	// Name service name
	Name types.Module = "Member"
)

// Module model
type Module struct {
	restHandler    *resthandler.RestHandler
	grpcHandler    *grpchandler.GRPCHandler
	graphqlHandler *graphqlhandler.GraphQLHandler

	workerHandlers map[types.Worker]interfaces.WorkerHandler
}

// NewModule module constructor
func NewModule(deps dependency.Dependency) *Module {
	var mod Module
	mod.restHandler = resthandler.NewRestHandler(deps.GetMiddleware())
	mod.grpcHandler = grpchandler.NewGRPCHandler(deps.GetMiddleware())
	mod.graphqlHandler = graphqlhandler.NewGraphQLHandler(deps.GetMiddleware())

	mod.workerHandlers = map[types.Worker]interfaces.WorkerHandler{
		// types.Kafka: workerhandler.NewKafkaHandler(),
	}
	return &mod
}

// RestHandler method
func (m *Module) RestHandler() interfaces.EchoRestHandler {
	return m.restHandler
}

// GRPCHandler method
func (m *Module) GRPCHandler() interfaces.GRPCHandler {
	return m.grpcHandler
}

// GraphQLHandler method
func (m *Module) GraphQLHandler() interfaces.GraphQLHandler {
	return nil
}

// WorkerHandler method
func (m *Module) WorkerHandler(workerType types.Worker) interfaces.WorkerHandler {
	return m.workerHandlers[workerType]
}

// Name get module name
func (m *Module) Name() types.Module {
	return Name
}
