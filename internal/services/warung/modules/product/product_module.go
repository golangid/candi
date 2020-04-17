package product

import (
	"agungdwiprasetyo.com/backend-microservices/internal/factory/base"
	"agungdwiprasetyo.com/backend-microservices/internal/factory/constant"
	"agungdwiprasetyo.com/backend-microservices/internal/factory/interfaces"
	"agungdwiprasetyo.com/backend-microservices/internal/services/warung/modules/product/delivery"
	"agungdwiprasetyo.com/backend-microservices/pkg/helper"
)

const (
	// Product service name
	Product constant.Module = iota
)

// Module model
type Module struct {
	restHandler    *delivery.RestProductHandler
	graphqlHandler *delivery.GraphQLHandler
}

// NewModule module constructor
func NewModule(params *base.ModuleParam) *Module {

	var mod Module
	mod.restHandler = delivery.NewRestProductHandler(params.Middleware)
	mod.graphqlHandler = delivery.NewGraphQLHandler(params.Middleware)
	return &mod
}

// RestHandler method
func (m *Module) RestHandler(version string) (d interfaces.EchoRestHandler) {
	switch version {
	case helper.V1:
		d = m.restHandler
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
	return "Product", m.graphqlHandler
}

// SubscriberHandler method
func (m *Module) SubscriberHandler(subsType constant.Subscriber) interfaces.SubscriberHandler {
	return nil
}

// Name get module name
func (m *Module) Name() constant.Module {
	return Product
}
