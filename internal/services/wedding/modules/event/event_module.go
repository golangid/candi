package event

import (
	"agungdwiprasetyo.com/backend-microservices/internal/services/wedding/modules/event/delivery"
	"agungdwiprasetyo.com/backend-microservices/internal/services/wedding/modules/event/repository"
	"agungdwiprasetyo.com/backend-microservices/internal/services/wedding/modules/event/usecase"
	"agungdwiprasetyo.com/backend-microservices/pkg/codebase/factory/base"
	"agungdwiprasetyo.com/backend-microservices/pkg/codebase/factory/constant"
	"agungdwiprasetyo.com/backend-microservices/pkg/codebase/factory/interfaces"
)

const (
	// Event service name
	Event constant.Module = "Event"
)

// Module model
type Module struct {
	graphqlHandler *delivery.GraphQLHandler
	kafkaHandler   *delivery.KafkaHandler
}

// NewModule module constructor
func NewModule(deps *base.Dependency) *Module {
	repo := repository.NewRepositoryMongo(deps.Config.MongoRead, deps.Config.MongoWrite)
	uc := usecase.NewEventUsecase(repo)

	var mod Module
	mod.graphqlHandler = delivery.NewGraphQLHandler(deps.Middleware, uc)
	mod.kafkaHandler = delivery.NewKafkaHandler([]string{
		"event",
	})
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
	return string(Event), m.graphqlHandler
}

// WorkerHandler method
func (m *Module) WorkerHandler(workerType constant.Worker) interfaces.WorkerHandler {
	switch workerType {
	case constant.Kafka:
		return m.kafkaHandler
	default:
		return nil
	}
}

// Name get module name
func (m *Module) Name() constant.Module {
	return Event
}
