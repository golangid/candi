package customer

import (
	"agungdwiprasetyo.com/backend-microservices/internal/factory/base"
	"agungdwiprasetyo.com/backend-microservices/internal/factory/constant"
	"agungdwiprasetyo.com/backend-microservices/internal/factory/interfaces"
	"agungdwiprasetyo.com/backend-microservices/internal/services/user-service/modules/customer/delivery/graphqlhandler"
	"agungdwiprasetyo.com/backend-microservices/internal/services/user-service/modules/customer/delivery/grpchandler"
	"agungdwiprasetyo.com/backend-microservices/internal/services/user-service/modules/customer/delivery/resthandler"
	"agungdwiprasetyo.com/backend-microservices/internal/services/user-service/modules/customer/delivery/workerhandler"
)

const (
	// Name service name
	Name constant.Module = "Customer"
)

// Module model
type Module struct {
	restHandler    *resthandler.RestHandler
	grpcHandler    *grpchandler.GRPCHandler
	graphqlHandler *graphqlhandler.GraphQLHandler
	kafkaHandler   *workerhandler.KafkaHandler
}

// NewModule module constructor
func NewModule(deps *base.Dependency) *Module {
	var mod Module
	mod.restHandler = resthandler.NewRestHandler(deps.Middleware)
	mod.grpcHandler = grpchandler.NewGRPCHandler(deps.Middleware)
	mod.graphqlHandler = graphqlhandler.NewGraphQLHandler(deps.Middleware)
	mod.kafkaHandler = workerhandler.NewKafkaHandler([]string{"test"})
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
func (m *Module) GraphQLHandler() (name string, resolver interface{}) {
	return string(Name), m.graphqlHandler
}

// WorkerHandler method
func (m *Module) WorkerHandler(workerType constant.Worker) interfaces.WorkerHandler {
	switch workerType {
	case constant.Kafka:
		return m.kafkaHandler
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

