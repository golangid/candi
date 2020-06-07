package main

const moduleMainTemplate = `package {{$.module}}

import (
	"{{$.PackageName}}/internal/factory/base"
	"{{$.PackageName}}/internal/factory/constant"
	"{{$.PackageName}}/internal/factory/interfaces"
	"{{$.PackageName}}/pkg/helper"
)

const (
	// Name service name
	Name constant.Module = "{{$.module}}"
)

// Module model
type Module struct {
}

// NewModule module constructor
func NewModule(params *base.ModuleParam) *Module {
	var mod Module
	return &mod
}

// RestHandler method
func (m *Module) RestHandler(version string) (d interfaces.EchoRestHandler) {
	switch version {
	case helper.V1:
		d = nil
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
		return nil
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
