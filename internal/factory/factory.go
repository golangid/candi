package factory

import (
	"github.com/agungdwiprasetyo/backend-microservices/internal/factory/constant"
	"github.com/agungdwiprasetyo/backend-microservices/internal/factory/interfaces"
)

// ServiceFactory factory
type ServiceFactory interface {
	Modules() []ModuleFactory
	Name() constant.Service
}

// ModuleFactory factory
type ModuleFactory interface {
	RestHandler(version string) interfaces.EchoRestDelivery
	GRPCHandler() interfaces.GRPCDelivery
	SubscriberHandler(subsType constant.Subscriber) interfaces.SubscriberDelivery
	Name() constant.Module
}
