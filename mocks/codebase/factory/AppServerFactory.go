// Code generated by mockery v2.49.1. DO NOT EDIT.

package mocks

import (
	context "context"

	mock "github.com/stretchr/testify/mock"
)

// AppServerFactory is an autogenerated mock type for the AppServerFactory type
type AppServerFactory struct {
	mock.Mock
}

// Name provides a mock function with given fields:
func (_m *AppServerFactory) Name() string {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Name")
	}

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// Serve provides a mock function with given fields:
func (_m *AppServerFactory) Serve() {
	_m.Called()
}

// Shutdown provides a mock function with given fields: ctx
func (_m *AppServerFactory) Shutdown(ctx context.Context) {
	_m.Called(ctx)
}

// NewAppServerFactory creates a new instance of AppServerFactory. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewAppServerFactory(t interface {
	mock.TestingT
	Cleanup(func())
}) *AppServerFactory {
	mock := &AppServerFactory{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
