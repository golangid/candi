package event

import (
	"github.com/agungdwiprasetyo/backend-microservices/internal/factory/base"
	"github.com/agungdwiprasetyo/backend-microservices/internal/factory/constant"
	"github.com/agungdwiprasetyo/backend-microservices/internal/factory/interfaces"
	"github.com/agungdwiprasetyo/backend-microservices/internal/services/wedding/modules/event/delivery"
	"github.com/agungdwiprasetyo/backend-microservices/internal/services/wedding/modules/event/repository"
	"github.com/agungdwiprasetyo/backend-microservices/internal/services/wedding/modules/event/usecase"
	"github.com/agungdwiprasetyo/backend-microservices/pkg/helper"
)

const (
	// Event service name
	Event constant.Module = iota
)

// Module model
type Module struct {
	graphqlHandler *delivery.GraphQLHandler
}

// NewModule module constructor
func NewModule(params *base.ModuleParam) *Module {
	repo := repository.NewRepositoryMongo(params.Config.MongoRead, params.Config.MongoWrite)
	uc := usecase.NewEventUsecase(repo)

	var mod Module
	mod.graphqlHandler = delivery.NewGraphQLHandler(params.Middleware, uc)
	return &mod
}

// RestHandler method
func (m *Module) RestHandler(version string) (d interfaces.EchoRestDelivery) {
	switch version {
	case helper.V1:
		d = nil
	case helper.V2:
		d = nil // TODO versioning
	}
	return
}

// GRPCHandler method
func (m *Module) GRPCHandler() interfaces.GRPCDelivery {
	return nil
}

// GraphQLHandler method
func (m *Module) GraphQLHandler() (name string, resolver interface{}) {
	return "Event", m.graphqlHandler
}

// SubscriberHandler method
func (m *Module) SubscriberHandler(subsType constant.Subscriber) interfaces.SubscriberDelivery {
	return nil
}

// Name get module name
func (m *Module) Name() constant.Module {
	return Event
}
