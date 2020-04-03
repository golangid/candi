package factory

import (
	"github.com/agungdwiprasetyo/backend-microservices/config"
	"github.com/agungdwiprasetyo/backend-microservices/internal/factory/base"
	"github.com/agungdwiprasetyo/backend-microservices/internal/factory/constant"
	"github.com/agungdwiprasetyo/backend-microservices/internal/factory/interfaces"
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
