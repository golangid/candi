package invitation

import (
	"github.com/agungdwiprasetyo/backend-microservices/internal/factory/base"
	"github.com/agungdwiprasetyo/backend-microservices/internal/factory/constant"
	"github.com/agungdwiprasetyo/backend-microservices/internal/factory/interfaces"
	"github.com/agungdwiprasetyo/backend-microservices/internal/services/wedding/modules/invitation/delivery"
	"github.com/agungdwiprasetyo/backend-microservices/pkg/helper"
)

const (
	// Invitation service name
	Invitation constant.Module = iota
)

// Module model
type Module struct {
	restHandler    *delivery.RestInvitationHandler
	graphqlHandler *delivery.GraphQLHandler
	kafkaHandler   *delivery.KafkaHandler
}

// NewModule module constructor
func NewModule(params *base.ModuleParam) *Module {

	var mod Module
	mod.restHandler = delivery.NewRestInvitationHandler(params.Middleware)
	mod.graphqlHandler = delivery.NewGraphQLHandler(params.Middleware)
	mod.kafkaHandler = delivery.NewKafkaHandler([]string{"test"})
	return &mod
}

// RestHandler method
func (m *Module) RestHandler(version string) (d interfaces.EchoRestDelivery) {
	switch version {
	case helper.V1:
		d = m.restHandler
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
	return "Invitation", m.graphqlHandler
}

// SubscriberHandler method
func (m *Module) SubscriberHandler(subsType constant.Subscriber) interfaces.SubscriberDelivery {
	switch subsType {
	case constant.Kafka:
		return m.kafkaHandler
	}
	return nil
}

// Name get module name
func (m *Module) Name() constant.Module {
	return Invitation
}
