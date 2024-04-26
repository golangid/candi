// Code generated by mockery v2.42.1. DO NOT EDIT.

package mocks

import (
	tracer "github.com/golangid/candi/tracer"
	mock "github.com/stretchr/testify/mock"
)

// OptionFunc is an autogenerated mock type for the OptionFunc type
type OptionFunc struct {
	mock.Mock
}

// Execute provides a mock function with given fields: _a0
func (_m *OptionFunc) Execute(_a0 *tracer.Option) {
	_m.Called(_a0)
}

// NewOptionFunc creates a new instance of OptionFunc. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewOptionFunc(t interface {
	mock.TestingT
	Cleanup(func())
}) *OptionFunc {
	mock := &OptionFunc{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
