package main

const moduleMainTemplate = `package {{clean $.module}}

import (
	"{{$.PackageName}}/internal/{{.ServiceName}}/modules/{{$.module}}/delivery/graphqlhandler"
	"{{$.PackageName}}/internal/{{.ServiceName}}/modules/{{$.module}}/delivery/grpchandler"
	"{{$.PackageName}}/internal/{{.ServiceName}}/modules/{{$.module}}/delivery/resthandler"
	"{{$.PackageName}}/internal/{{.ServiceName}}/modules/{{$.module}}/delivery/workerhandler"
	"{{$.PackageName}}/pkg/codebase/factory/dependency"
	"{{$.PackageName}}/pkg/codebase/factory/constant"
	"{{$.PackageName}}/pkg/codebase/interfaces"
)

const (
	// Name service name
	Name constant.Module = "{{clean (upper $.module)}}"
)

// Module model
type Module struct {
	restHandler    *resthandler.RestHandler
	grpcHandler    *grpchandler.GRPCHandler
	graphqlHandler *graphqlhandler.GraphQLHandler

	workerHandlers map[constant.Worker]interfaces.WorkerHandler
}

// NewModule module constructor
func NewModule(deps dependency.Dependency) *Module {
	var mod Module
	mod.restHandler = resthandler.NewRestHandler(deps.GetMiddleware())
	mod.grpcHandler = grpchandler.NewGRPCHandler(deps.GetMiddleware())
	mod.graphqlHandler = graphqlhandler.NewGraphQLHandler(deps.GetMiddleware())

	mod.workerHandlers = map[constant.Worker]interfaces.WorkerHandler{
		constant.Kafka: workerhandler.NewKafkaHandler(), // example worker
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
