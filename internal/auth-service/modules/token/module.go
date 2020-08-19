package token

import (
	"agungdwiprasetyo.com/backend-microservices/internal/auth-service/modules/token/delivery/graphqlhandler"
	"agungdwiprasetyo.com/backend-microservices/internal/auth-service/modules/token/delivery/grpchandler"
	"agungdwiprasetyo.com/backend-microservices/internal/auth-service/modules/token/delivery/resthandler"
	"agungdwiprasetyo.com/backend-microservices/internal/auth-service/modules/token/delivery/workerhandler"
	"agungdwiprasetyo.com/backend-microservices/internal/auth-service/modules/token/usecase"
	"agungdwiprasetyo.com/backend-microservices/pkg/codebase/factory/dependency"
	"agungdwiprasetyo.com/backend-microservices/pkg/codebase/factory/types"
	"agungdwiprasetyo.com/backend-microservices/pkg/codebase/interfaces"
)

const (
	// Name service name
	Name types.Module = "Token"
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
	uc := usecase.NewTokenUsecase(deps.GetKey().PublicKey(), deps.GetKey().PrivateKey())

	var mod Module
	mod.restHandler = resthandler.NewRestHandler(deps.GetMiddleware())
	mod.grpcHandler = grpchandler.NewGRPCHandler(deps.GetMiddleware(), uc)
	mod.graphqlHandler = graphqlhandler.NewGraphQLHandler(deps.GetMiddleware(), uc)

	mod.workerHandlers = map[types.Worker]interfaces.WorkerHandler{
		types.Kafka: workerhandler.NewKafkaHandler(), // example worker
		// add more worker type from delivery, implement "interfaces.WorkerHandler"
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
	return m.graphqlHandler
}

// WorkerHandler method
func (m *Module) WorkerHandler(workerType types.Worker) interfaces.WorkerHandler {
	return m.workerHandlers[workerType]
}

// Name get module name
func (m *Module) Name() types.Module {
	return Name
}
