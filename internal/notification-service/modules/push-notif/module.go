package pushnotif

import (
	"agungdwiprasetyo.com/backend-microservices/internal/notification-service/modules/push-notif/delivery/graphqlhandler"
	"agungdwiprasetyo.com/backend-microservices/internal/notification-service/modules/push-notif/delivery/grpchandler"
	"agungdwiprasetyo.com/backend-microservices/internal/notification-service/modules/push-notif/delivery/resthandler"
	"agungdwiprasetyo.com/backend-microservices/internal/notification-service/modules/push-notif/delivery/workerhandler"
	"agungdwiprasetyo.com/backend-microservices/internal/notification-service/modules/push-notif/repository"
	"agungdwiprasetyo.com/backend-microservices/internal/notification-service/modules/push-notif/usecase"
	"agungdwiprasetyo.com/backend-microservices/pkg/codebase/factory/constant"
	"agungdwiprasetyo.com/backend-microservices/pkg/codebase/factory/dependency"
	"agungdwiprasetyo.com/backend-microservices/pkg/codebase/interfaces"
)

const (
	// Name service name
	Name constant.Module = "PushNotif"
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
	repo := repository.NewRepository()
	uc := usecase.NewPushNotifUsecase(repo)

	var mod Module
	mod.restHandler = resthandler.NewRestHandler(deps.GetMiddleware(), uc)
	mod.grpcHandler = grpchandler.NewGRPCHandler(deps.GetMiddleware())
	mod.graphqlHandler = graphqlhandler.NewGraphQLHandler(deps.GetMiddleware(), uc)

	mod.workerHandlers = map[constant.Worker]interfaces.WorkerHandler{
		constant.Kafka: workerhandler.NewKafkaHandler(uc), // example worker
		// add more worker type from delivery, implement "interfaces.WorkerHandler"
		constant.Scheduler:       workerhandler.NewCronHandler(uc),
		constant.RedisSubscriber: workerhandler.NewRedisHandler(Name, uc),
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
