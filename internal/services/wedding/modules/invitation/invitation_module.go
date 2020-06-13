package invitation

import (
	"agungdwiprasetyo.com/backend-microservices/internal/factory/base"
	"agungdwiprasetyo.com/backend-microservices/internal/factory/constant"
	"agungdwiprasetyo.com/backend-microservices/internal/factory/interfaces"
	"agungdwiprasetyo.com/backend-microservices/internal/services/wedding/modules/invitation/delivery"
)

const (
	// Invitation service name
	Invitation constant.Module = "Invitation"
)

// Module model
type Module struct {
	restHandler    *delivery.RestInvitationHandler
	graphqlHandler *delivery.GraphQLHandler
	kafkaHandler   *delivery.KafkaHandler
}

// NewModule module constructor
func NewModule(deps *base.Dependency) *Module {

	var mod Module
	mod.restHandler = delivery.NewRestInvitationHandler(deps.Middleware)
	mod.graphqlHandler = delivery.NewGraphQLHandler(deps.Middleware)
	mod.kafkaHandler = delivery.NewKafkaHandler([]string{"test", "coba"})
	return &mod
}

// RestHandler method
func (m *Module) RestHandler() interfaces.EchoRestHandler {
	return m.restHandler
}

// GRPCHandler method
func (m *Module) GRPCHandler() interfaces.GRPCHandler {
	return nil
}

// GraphQLHandler method
func (m *Module) GraphQLHandler() (name string, resolver interface{}) {
	return string(Invitation), m.graphqlHandler
}

// WorkerHandler method
func (m *Module) WorkerHandler(workerType constant.Worker) interfaces.WorkerHandler {
	switch workerType {
	case constant.Kafka:
		return m.kafkaHandler
	}
	return nil
}

// Name get module name
func (m *Module) Name() constant.Module {
	return Invitation
}
