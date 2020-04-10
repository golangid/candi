package factory

import (
	"agungdwiprasetyo.com/backend-microservices/config"
	"agungdwiprasetyo.com/backend-microservices/internal/factory/base"
	"agungdwiprasetyo.com/backend-microservices/internal/factory/constant"
	"agungdwiprasetyo.com/backend-microservices/internal/factory/interfaces"
)

// ServiceFactory factory
type ServiceFactory interface {
	GetConfig() *config.Config
	Modules(params *base.ModuleParam) []ModuleFactory
	Name() constant.Service
}

// ModuleFactory factory
type ModuleFactory interface {
	RestHandler(version string) interfaces.EchoRestDelivery
	GRPCHandler() interfaces.GRPCDelivery
	GraphQLHandler() (name string, resolver interface{})
	SubscriberHandler(subsType constant.Subscriber) interfaces.SubscriberDelivery
	Name() constant.Module
}
