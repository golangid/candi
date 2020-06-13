package main

const moduleMainTemplate = `package {{clean $.module}}

import (
	"{{$.PackageName}}/internal/factory/base"
	"{{$.PackageName}}/internal/factory/constant"
	"{{$.PackageName}}/internal/factory/interfaces"
	"{{$.PackageName}}/internal/services/coba/modules/{{clean $.module}}/delivery/graphqlhandler"
	"{{$.PackageName}}/internal/services/coba/modules/{{clean $.module}}/delivery/resthandler"
	"{{$.PackageName}}/internal/services/coba/modules/{{clean $.module}}/delivery/workerhandler"
)

const (
	// Name service name
	Name constant.Module = "{{upper $.module}}"
)

// Module model
type Module struct {
	restHandler    *resthandler.RestHandler
	graphqlHandler *graphqlhandler.GraphQLHandler
	kafkaHandler   *workerhandler.KafkaHandler
}

// NewModule module constructor
func NewModule(deps *base.Dependency) *Module {
	var mod Module
	mod.restHandler = resthandler.NewRestHandler(deps.Middleware)
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
	return nil
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

`

const defaultFile = `package {{$.packageName}}`

func defaultDataSource(fileName string) []byte {
	return loadTemplate(defaultFile, map[string]string{"packageName": fileName})
}
