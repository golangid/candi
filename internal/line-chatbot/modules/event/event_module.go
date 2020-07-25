package event

import (
	"agungdwiprasetyo.com/backend-microservices/internal/line-chatbot/modules/event/delivery/graphqlhandler"
	"agungdwiprasetyo.com/backend-microservices/internal/line-chatbot/modules/event/delivery/grpchandler"
	"agungdwiprasetyo.com/backend-microservices/internal/line-chatbot/modules/event/repository"
	"agungdwiprasetyo.com/backend-microservices/internal/line-chatbot/modules/event/usecase"
	"agungdwiprasetyo.com/backend-microservices/pkg/codebase/factory/dependency"
	"agungdwiprasetyo.com/backend-microservices/pkg/codebase/factory/types"
	"agungdwiprasetyo.com/backend-microservices/pkg/codebase/interfaces"
)

const (
	// Event service name
	Event types.Module = "Event"
)

// Module model
type Module struct {
	graphqlHandler *graphqlhandler.GraphQLHandler
	grpcHandler    *grpchandler.GRPCHandler
}

// NewModule module constructor
func NewModule(deps dependency.Dependency) *Module {
	repo := repository.NewRepoMongo(deps.GetMongoDatabase().WriteDB())
	uc := usecase.NewEventUsecase(repo)

	var mod Module
	mod.graphqlHandler = graphqlhandler.NewGraphQLHandler(string(Event), deps.GetMiddleware(), uc)
	mod.grpcHandler = grpchandler.NewGRPCHandler(uc)
	return &mod
}

// RestHandler method
func (m *Module) RestHandler() interfaces.EchoRestHandler {
	return nil
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
	return nil
}

// Name get module name
func (m *Module) Name() types.Module {
	return Event
}
