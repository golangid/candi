// Code generated by mockery v2.42.1. DO NOT EDIT.

package mocks

import (
	types "github.com/golangid/candi/codebase/factory/types"
	mock "github.com/stretchr/testify/mock"
)

// WorkerHandler is an autogenerated mock type for the WorkerHandler type
type WorkerHandler struct {
	mock.Mock
}

// MountHandlers provides a mock function with given fields: group
func (_m *WorkerHandler) MountHandlers(group *types.WorkerHandlerGroup) {
	_m.Called(group)
}

// NewWorkerHandler creates a new instance of WorkerHandler. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewWorkerHandler(t interface {
	mock.TestingT
	Cleanup(func())
}) *WorkerHandler {
	mock := &WorkerHandler{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
