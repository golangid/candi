package public

import (
	"agungdwiprasetyo.com/backend-microservices/internal/factory/base"
	"agungdwiprasetyo.com/backend-microservices/internal/factory/constant"
	"agungdwiprasetyo.com/backend-microservices/internal/factory/interfaces"
	"agungdwiprasetyo.com/backend-microservices/internal/services/cms/modules/public/delivery/graphqlhandler"
	"agungdwiprasetyo.com/backend-microservices/internal/services/cms/modules/public/repository"
	"agungdwiprasetyo.com/backend-microservices/internal/services/cms/modules/public/usecase"
	"agungdwiprasetyo.com/backend-microservices/pkg/helper"
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
func NewModule(params *base.ModuleParam) *Module {
	repo := repository.NewRepoMongo(params.Config.MongoWrite)
	uc := usecase.NewPublicUsecase(repo)

	var mod Module
	mod.graphqlHandler = graphqlhandler.NewGraphQLHandler(params.Middleware, uc)
	return &mod
}

// RestHandler method
func (m *Module) RestHandler(version string) (d interfaces.EchoRestHandler) {
	switch version {
	case helper.V1:
		d = nil
	case helper.V2:
		d = nil
	}
	return
}

// GRPCHandler method
func (m *Module) GRPCHandler() interfaces.GRPCHandler {
	return nil
}

// GraphQLHandler method
func (m *Module) GraphQLHandler() (name string, resolver interface{}) {
	return string(Name), m.graphqlHandler
}

// SubscriberHandler method
func (m *Module) SubscriberHandler(subsType constant.Subscriber) interfaces.SubscriberHandler {
	switch subsType {
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
