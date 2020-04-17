package event

import (
	"agungdwiprasetyo.com/backend-microservices/internal/factory/base"
	"agungdwiprasetyo.com/backend-microservices/internal/factory/constant"
	"agungdwiprasetyo.com/backend-microservices/internal/factory/interfaces"
	"agungdwiprasetyo.com/backend-microservices/internal/services/wedding/modules/event/delivery"
	"agungdwiprasetyo.com/backend-microservices/internal/services/wedding/modules/event/repository"
	"agungdwiprasetyo.com/backend-microservices/internal/services/wedding/modules/event/usecase"
	"agungdwiprasetyo.com/backend-microservices/pkg/helper"
)

const (
	// Event service name
	Event constant.Module = iota
)

// Module model
type Module struct {
	graphqlHandler *delivery.GraphQLHandler
	kafkaHandler   *delivery.KafkaHandler
}

// NewModule module constructor
func NewModule(params *base.ModuleParam) *Module {
	repo := repository.NewRepositoryMongo(params.Config.MongoRead, params.Config.MongoWrite)
	uc := usecase.NewEventUsecase(repo)

	var mod Module
	mod.graphqlHandler = delivery.NewGraphQLHandler(params.Middleware, uc)
	mod.kafkaHandler = delivery.NewKafkaHandler([]string{
		"coba",
	})
	return &mod
}

// RestHandler method
func (m *Module) RestHandler(version string) (d interfaces.EchoRestHandler) {
	switch version {
	case helper.V1:
		d = nil
	case helper.V2:
		d = nil // TODO versioning
	}
	return
}

// GRPCHandler method
func (m *Module) GRPCHandler() interfaces.GRPCHandler {
	return nil
}

// GraphQLHandler method
func (m *Module) GraphQLHandler() (name string, resolver interface{}) {
	return "Event", m.graphqlHandler
}

// SubscriberHandler method
func (m *Module) SubscriberHandler(subsType constant.Subscriber) interfaces.SubscriberHandler {
	switch subsType {
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
