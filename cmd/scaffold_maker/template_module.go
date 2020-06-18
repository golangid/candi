package main

const moduleMainTemplate = `package {{clean $.module}}

import (
	"{{$.PackageName}}/internal/{{.ServiceName}}/modules/{{clean $.module}}/delivery/graphqlhandler"
	"{{$.PackageName}}/internal/{{.ServiceName}}/modules/{{clean $.module}}/delivery/grpchandler"
	"{{$.PackageName}}/internal/{{.ServiceName}}/modules/{{clean $.module}}/delivery/resthandler"
	"{{$.PackageName}}/internal/{{.ServiceName}}/modules/{{clean $.module}}/delivery/workerhandler"
	"{{$.PackageName}}/pkg/codebase/factory/base"
	"{{$.PackageName}}/pkg/codebase/factory/constant"
	"{{$.PackageName}}/pkg/codebase/interfaces"
)

const (
	// Name service name
	Name constant.Module = "{{upper $.module}}"
)

// Module model
type Module struct {
	restHandler    *resthandler.RestHandler
	grpcHandler    *grpchandler.GRPCHandler
	graphqlHandler *graphqlhandler.GraphQLHandler

	workerHandlers map[constant.Worker]interfaces.WorkerHandler
}

// NewModule module constructor
func NewModule(deps *base.Dependency) *Module {
	var mod Module
	mod.restHandler = resthandler.NewRestHandler(deps.Middleware)
	mod.grpcHandler = grpchandler.NewGRPCHandler(deps.Middleware)
	mod.graphqlHandler = graphqlhandler.NewGraphQLHandler(deps.Middleware)

	mod.workerHandlers = map[constant.Worker]interfaces.WorkerHandler{
		constant.Kafka: workerhandler.NewKafkaHandler([]string{"test"}), // example worker
		// add more worker type from delivery, implement "interfaces.WorkerHandler"
	}

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
	return m.workerHandlers[workerType]
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
