package product

import (
	"agungdwiprasetyo.com/backend-microservices/internal/services/warung/modules/product/delivery"
	"agungdwiprasetyo.com/backend-microservices/pkg/codebase/factory/base"
	"agungdwiprasetyo.com/backend-microservices/pkg/codebase/factory/constant"
	"agungdwiprasetyo.com/backend-microservices/pkg/codebase/factory/interfaces"
)

const (
	// Product service name
	Product constant.Module = "Product"
)

// Module model
type Module struct {
	restHandler    *delivery.RestProductHandler
	graphqlHandler *delivery.GraphQLHandler
}

// NewModule module constructor
func NewModule(deps *base.Dependency) *Module {

	var mod Module
	mod.restHandler = delivery.NewRestProductHandler(deps.Middleware)
	mod.graphqlHandler = delivery.NewGraphQLHandler(deps.Middleware)
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
	return string(Product), m.graphqlHandler
}

// WorkerHandler method
func (m *Module) WorkerHandler(workerType constant.Worker) interfaces.WorkerHandler {
	return nil
}

// Name get module name
func (m *Module) Name() constant.Module {
	return Product
}
