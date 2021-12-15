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
	moduleName types.Module = "{{upper (camel .ModuleName)}}"
)

// Module model
type Module struct {
	restHandler    interfaces.RESTHandler
	grpcHandler    interfaces.GRPCHandler
	graphqlHandler interfaces.GraphQLHandler

	workerHandlers map[types.Worker]interfaces.WorkerHandler
	serverHandlers map[types.Server]interfaces.ServerHandler
}

// NewModule module constructor
func NewModule(deps dependency.Dependency) *Module {
	var mod Module
	{{if not .RestHandler}}// {{end}}mod.restHandler = resthandler.NewRestHandler(usecase.GetSharedUsecase(), deps)
	{{if not .GRPCHandler}}// {{end}}mod.grpcHandler = grpchandler.NewGRPCHandler(usecase.GetSharedUsecase(), deps)
	{{if not .GraphQLHandler}}// {{end}}mod.graphqlHandler = graphqlhandler.NewGraphQLHandler(usecase.GetSharedUsecase(), deps)

	mod.workerHandlers = map[types.Worker]interfaces.WorkerHandler{
		{{if not .KafkaHandler}}// {{end}}types.Kafka:           workerhandler.NewKafkaHandler(usecase.GetSharedUsecase(), deps),
		{{if not .SchedulerHandler}}// {{end}}types.Scheduler:       workerhandler.NewCronHandler(usecase.GetSharedUsecase(), deps),
		{{if not .RedisSubsHandler}}// {{end}}types.RedisSubscriber: workerhandler.NewRedisHandler(usecase.GetSharedUsecase(), deps),
		{{if not .TaskQueueHandler}}// {{end}}types.TaskQueue:       workerhandler.NewTaskQueueHandler(usecase.GetSharedUsecase(), deps),
		{{if not .PostgresListenerHandler}}// {{end}}types.PostgresListener: workerhandler.NewPostgresListenerHandler(usecase.GetSharedUsecase(), deps),
		{{if not .RabbitMQHandler}}// {{end}}types.RabbitMQ: workerhandler.NewRabbitMQHandler(usecase.GetSharedUsecase(), deps),
	}

	mod.serverHandlers = map[types.Server]interfaces.ServerHandler{
		// 
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

// ServerHandler additional server type (another rest framework, p2p, and many more)
func (m *Module) ServerHandler(serverType types.Server) interfaces.ServerHandler {
	return m.serverHandlers[serverType]
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
