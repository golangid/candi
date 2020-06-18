package event

import (
	"agungdwiprasetyo.com/backend-microservices/internal/line-chatbot/modules/event/delivery/graphqlhandler"
	"agungdwiprasetyo.com/backend-microservices/internal/line-chatbot/modules/event/delivery/grpchandler"
	"agungdwiprasetyo.com/backend-microservices/internal/line-chatbot/modules/event/repository"
	"agungdwiprasetyo.com/backend-microservices/internal/line-chatbot/modules/event/usecase"
	"agungdwiprasetyo.com/backend-microservices/pkg/codebase/factory/base"
	"agungdwiprasetyo.com/backend-microservices/pkg/codebase/factory/constant"
	"agungdwiprasetyo.com/backend-microservices/pkg/codebase/interfaces"
)

const (
	// Event service name
	Event constant.Module = "Event"
)

// Module model
type Module struct {
	graphqlHandler *graphqlhandler.GraphQLHandler
	grpcHandler    *grpchandler.GRPCHandler
}

// NewModule module constructor
func NewModule(deps *base.Dependency) *Module {
	repo := repository.NewRepoMongo(deps.Config.MongoWrite)
	uc := usecase.NewEventUsecase(repo)

	var mod Module
	mod.graphqlHandler = graphqlhandler.NewGraphQLHandler(deps.Middleware, uc)
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
func (m *Module) GraphQLHandler() (name string, resolver interface{}) {
	return string(Event), m.graphqlHandler
}

// WorkerHandler method
func (m *Module) WorkerHandler(workerType constant.Worker) interfaces.WorkerHandler {
	return nil
}

// Name get module name
func (m *Module) Name() constant.Module {
	return Event
}
