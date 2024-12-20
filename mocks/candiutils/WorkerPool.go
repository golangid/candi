// Code generated by mockery v2.49.1. DO NOT EDIT.

package mocks

import (
	context "context"

	mock "github.com/stretchr/testify/mock"
)

// WorkerPool is an autogenerated mock type for the WorkerPool type
type WorkerPool[T any] struct {
	mock.Mock
}

// AddJob provides a mock function with given fields: job
func (_m *WorkerPool[T]) AddJob(job T) {
	_m.Called(job)
}

// Dispatch provides a mock function with given fields: ctx, jobFunc
func (_m *WorkerPool[T]) Dispatch(ctx context.Context, jobFunc func(context.Context, T)) {
	_m.Called(ctx, jobFunc)
}

// Finish provides a mock function with given fields:
func (_m *WorkerPool[T]) Finish() {
	_m.Called()
}

// NewWorkerPool creates a new instance of WorkerPool. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewWorkerPool[T any](t interface {
	mock.TestingT
	Cleanup(func())
}) *WorkerPool[T] {
	mock := &WorkerPool[T]{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
