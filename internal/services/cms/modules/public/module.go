package public

import (
	"agungdwiprasetyo.com/backend-microservices/internal/factory/base"
	"agungdwiprasetyo.com/backend-microservices/internal/factory/constant"
	"agungdwiprasetyo.com/backend-microservices/internal/factory/interfaces"
	"agungdwiprasetyo.com/backend-microservices/internal/services/cms/modules/public/delivery/graphqlhandler"
	"agungdwiprasetyo.com/backend-microservices/internal/services/cms/modules/public/repository"
	"agungdwiprasetyo.com/backend-microservices/internal/services/cms/modules/public/usecase"
)

const (
	// Name service name
	Name constant.Module = "Public"
)

// Module model
type Module struct {
	graphqlHandler *graphqlhandler.GraphQLHandler
}

// NewModule module constructor
func NewModule(deps *base.Dependency) *Module {
	repo := repository.NewRepoMongo(deps.Config.MongoWrite)
	uc := usecase.NewPublicUsecase(repo)

	var mod Module
	mod.graphqlHandler = graphqlhandler.NewGraphQLHandler(deps.Middleware, uc)
	return &mod
}

// RestHandler method
func (m *Module) RestHandler() interfaces.EchoRestHandler {
	return nil
}

// GRPCHandler method
func (m *Module) GRPCHandler() interfaces.GRPCHandler {
	return nil
}

// GraphQLHandler method
func (m *Module) GraphQLHandler() (name string, resolver interface{}) {
	return string(Name), m.graphqlHandler
}

// WorkerHandler method
func (m *Module) WorkerHandler(workerType constant.Worker) interfaces.WorkerHandler {
	switch workerType {
	case constant.Kafka:
		return nil
	case constant.Redis:
		return nil
	case constant.RabbitMQ:
		return nil
	default:
		return nil
	}
}

// Name get module name
func (m *Module) Name() constant.Module {
	return Name
}
