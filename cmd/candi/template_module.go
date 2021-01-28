package main

const moduleMainTemplate = `// {{.Header}}

package {{clean .ModuleName}}

import (
	{{if not .GraphQLHandler}}// {{end}}"{{.PackagePrefix}}/internal/modules/{{cleanPathModule .ModuleName}}/delivery/graphqlhandler"
	{{if not .GRPCHandler}}// {{end}}"{{.PackagePrefix}}/internal/modules/{{cleanPathModule .ModuleName}}/delivery/grpchandler"
	{{if not .RestHandler}}// {{end}}"{{.PackagePrefix}}/internal/modules/{{cleanPathModule .ModuleName}}/delivery/resthandler"
	{{if not .IsWorkerActive}}// {{end}}"{{.PackagePrefix}}/internal/modules/{{cleanPathModule .ModuleName}}/delivery/workerhandler"
	"{{.PackagePrefix}}/pkg/shared/usecase"

	"{{.LibraryName}}/codebase/factory/dependency"
	"{{.LibraryName}}/codebase/factory/types"
	"{{.LibraryName}}/codebase/interfaces"
)

const (
	moduleName types.Module = "{{clean (upper .ModuleName)}}"
)

// Module model
type Module struct {
	restHandler    interfaces.RESTHandler
	grpcHandler    interfaces.GRPCHandler
	graphqlHandler interfaces.GraphQLHandler

	workerHandlers map[types.Worker]interfaces.WorkerHandler
}

// NewModule module constructor
func NewModule(deps dependency.Dependency) *Module {
	usecaseUOW := usecase.GetSharedUsecase()

	var mod Module
	{{if not .RestHandler}}// {{end}}mod.restHandler = resthandler.NewRestHandler(deps.GetMiddleware(), usecaseUOW.{{clean (upper .ModuleName)}}(), deps.GetValidator())
	{{if not .GRPCHandler}}// {{end}}mod.grpcHandler = grpchandler.NewGRPCHandler(deps.GetMiddleware(), usecaseUOW.{{clean (upper .ModuleName)}}(), deps.GetValidator())
	{{if not .GraphQLHandler}}// {{end}}mod.graphqlHandler = graphqlhandler.NewGraphQLHandler(deps.GetMiddleware(), usecaseUOW.{{clean (upper .ModuleName)}}(), deps.GetValidator())

	mod.workerHandlers = map[types.Worker]interfaces.WorkerHandler{
		{{if not .KafkaHandler}}// {{end}}types.Kafka:           workerhandler.NewKafkaHandler(usecaseUOW.{{clean (upper .ModuleName)}}(), deps.GetValidator()),
		{{if not .SchedulerHandler}}// {{end}}types.Scheduler:       workerhandler.NewCronHandler(usecaseUOW.{{clean (upper .ModuleName)}}(), deps.GetValidator()),
		{{if not .RedisSubsHandler}}// {{end}}types.RedisSubscriber: workerhandler.NewRedisHandler(usecaseUOW.{{clean (upper .ModuleName)}}(), deps.GetValidator()),
		{{if not .TaskQueueHandler}}// {{end}}types.TaskQueue:       workerhandler.NewTaskQueueHandler(usecaseUOW.{{clean (upper .ModuleName)}}(), deps.GetValidator()),
	}

	return &mod
}

// RESTHandler method
func (m *Module) RESTHandler() interfaces.RESTHandler {
	return m.restHandler
}

// GRPCHandler method
func (m *Module) GRPCHandler() interfaces.GRPCHandler {
	return m.grpcHandler
}

// GraphQLHandler method
func (m *Module) GraphQLHandler() interfaces.GraphQLHandler {
	return m.graphqlHandler
}

// WorkerHandler method
func (m *Module) WorkerHandler(workerType types.Worker) interfaces.WorkerHandler {
	return m.workerHandlers[workerType]
}

// Name get module name
func (m *Module) Name() types.Module {
	return moduleName
}
`

const defaultFile = `package {{$.packageName}}`

func defaultDataSource(fileName string) []byte {
	return loadTemplate(defaultFile, map[string]string{"packageName": fileName})
}
