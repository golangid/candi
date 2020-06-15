package auth

import (
	"agungdwiprasetyo.com/backend-microservices/internal/user-service/modules/auth/delivery/graphqlhandler"
	"agungdwiprasetyo.com/backend-microservices/internal/user-service/modules/auth/delivery/grpchandler"
	"agungdwiprasetyo.com/backend-microservices/internal/user-service/modules/auth/delivery/resthandler"
	"agungdwiprasetyo.com/backend-microservices/internal/user-service/modules/auth/delivery/workerhandler"
	"agungdwiprasetyo.com/backend-microservices/pkg/codebase/factory/base"
	"agungdwiprasetyo.com/backend-microservices/pkg/codebase/factory/constant"
	"agungdwiprasetyo.com/backend-microservices/pkg/codebase/factory/interfaces"
)

const (
	// Name service name
	Name constant.Module = "Auth"
)

// Module model
type Module struct {
	restHandler    *resthandler.RestHandler
	grpcHandler    *grpchandler.GRPCHandler
	graphqlHandler *graphqlhandler.GraphQLHandler

	workerHandlers map[constant.Worker]interfaces.WorkerHandler
}

// NewModule module constructor
func NewModule(deps *base.Dependency) *Module {
	var mod Module
	mod.restHandler = resthandler.NewRestHandler(deps.Middleware)
	mod.grpcHandler = grpchandler.NewGRPCHandler(deps.Middleware)
	mod.graphqlHandler = graphqlhandler.NewGraphQLHandler(deps.Middleware)

	mod.workerHandlers = map[constant.Worker]interfaces.WorkerHandler{
		constant.Kafka: workerhandler.NewKafkaHandler([]string{"test", "update-member"}),
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
func (m *Module) GraphQLHandler() (name string, resolver interface{}) {
	return string(Name), m.graphqlHandler
}

// WorkerHandler method
func (m *Module) WorkerHandler(workerType constant.Worker) interfaces.WorkerHandler {
	return m.workerHandlers[workerType]
}

// Name get module name
func (m *Module) Name() constant.Module {
	return Name
}
