package storage

import (
	"agungdwiprasetyo.com/backend-microservices/internal/storage-service/modules/storage/delivery/grpchandler"
	"agungdwiprasetyo.com/backend-microservices/internal/storage-service/modules/storage/usecase"
	"agungdwiprasetyo.com/backend-microservices/pkg/codebase/factory/dependency"
	"agungdwiprasetyo.com/backend-microservices/pkg/codebase/factory/types"
	"agungdwiprasetyo.com/backend-microservices/pkg/codebase/interfaces"
)

const (
	// Name service name
	Name types.Module = "storage"
)

// Module model
type Module struct {
	grpchandler *grpchandler.GRPCHandler
}

// NewModule module constructor
func NewModule(deps dependency.Dependency) *Module {
	uc := usecase.NewStorageUsecase()

	var mod Module

	mod.grpchandler = grpchandler.NewGRPCHandler(uc, deps.GetMiddleware())
	return &mod
}

// RestHandler method
func (m *Module) RestHandler() interfaces.EchoRestHandler {
	return nil
}

// GRPCHandler method
func (m *Module) GRPCHandler() interfaces.GRPCHandler {
	return m.grpchandler
}

// GraphQLHandler method
func (m *Module) GraphQLHandler() interfaces.GraphQLHandler {
	return nil
}

// WorkerHandler method
func (m *Module) WorkerHandler(workerType types.Worker) interfaces.WorkerHandler {
	return nil
}

// Name get module name
func (m *Module) Name() types.Module {
	return Name
}
