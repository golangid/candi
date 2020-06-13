package main

const moduleMainTemplate = `package {{clean $.module}}

import (
	"{{$.PackageName}}/internal/factory/base"
	"{{$.PackageName}}/internal/factory/constant"
	"{{$.PackageName}}/internal/factory/interfaces"
	"{{$.PackageName}}/internal/services/coba/modules/{{clean $.module}}/delivery/resthandler"
	"{{$.PackageName}}/internal/services/coba/modules/{{clean $.module}}/delivery/subscriberhandler"
	"{{$.PackageName}}/pkg/helper"
)

const (
	// Name service name
	Name constant.Module = "{{$.module}}"
)

// Module model
type Module struct {
	restHandler *resthandler.RestHandler
	kafkaHandler *subscriberhandler.KafkaHandler
}

// NewModule module constructor
func NewModule(deps *base.Dependency) *Module {
	var mod Module
	mod.restHandler = resthandler.NewRestHandler(deps.Middleware)
	mod.kafkaHandler = subscriberhandler.NewKafkaHandler([]string{"test"})
	return &mod
}

// RestHandler method
func (m *Module) RestHandler(version string) (d interfaces.EchoRestHandler) {
	switch version {
	case helper.V1:
		d = m.restHandler
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
	return string(Name), nil
}

// SubscriberHandler method
func (m *Module) SubscriberHandler(subsType constant.Subscriber) interfaces.SubscriberHandler {
	switch subsType {
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

`

const defaultFile = `package {{$.packageName}}`

func defaultDataSource(fileName string) []byte {
	return loadTemplate(defaultFile, map[string]string{"packageName": fileName})
}
