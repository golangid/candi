package user

import (
	"agungdwiprasetyo.com/backend-microservices/internal/factory/base"
	"agungdwiprasetyo.com/backend-microservices/internal/factory/constant"
	"agungdwiprasetyo.com/backend-microservices/internal/factory/interfaces"
	"agungdwiprasetyo.com/backend-microservices/internal/services/warung/modules/user/delivery"
	"agungdwiprasetyo.com/backend-microservices/pkg/helper"
)

const (
	// User service name
	User constant.Module = "User"
)

// Module model
type Module struct {
	restHandler *delivery.RestUserHandler
}

// NewModule module constructor
func NewModule(deps *base.Dependency) *Module {

	var mod Module
	mod.restHandler = delivery.NewRestUserHandler(deps.Middleware)
	return &mod
}

// RestHandler method
func (m *Module) RestHandler(version string) (d interfaces.EchoRestHandler) {
	switch version {
	case helper.V1:
		d = m.restHandler
	case helper.V2:
		d = nil // TODO versioning
	}
	return
}

// GRPCHandler method
func (m *Module) GRPCHandler() interfaces.GRPCHandler {
	return nil
}

// GraphQLHandler method
func (m *Module) GraphQLHandler() (name string, resolver interface{}) {
	return string(User), nil
}

// WorkerHandler method
func (m *Module) WorkerHandler(workerType constant.Worker) interfaces.WorkerHandler {
	return nil
}

// Name get module name
func (m *Module) Name() constant.Module {
	return User
}
