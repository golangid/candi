package product

import (
	"github.com/agungdwiprasetyo/backend-microservices/internal/factory/base"
	"github.com/agungdwiprasetyo/backend-microservices/internal/factory/constant"
	"github.com/agungdwiprasetyo/backend-microservices/internal/factory/interfaces"
	"github.com/agungdwiprasetyo/backend-microservices/internal/services/warung/modules/product/delivery"
	"github.com/agungdwiprasetyo/backend-microservices/pkg/helper"
)

const (
	// Product service name
	Product constant.Module = iota
)

// Module model
type Module struct {
	restHandler *delivery.RestProductHandler
}

// NewModule module constructor
func NewModule(params *base.ModuleParam) *Module {

	var mod Module
	mod.restHandler = delivery.NewRestProductHandler(params.Middleware)
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

// SubscriberHandler method
func (m *Module) SubscriberHandler(subsType constant.Subscriber) interfaces.SubscriberDelivery {
	return nil
}

// Name get module name
func (m *Module) Name() constant.Module {
	return Product
}
