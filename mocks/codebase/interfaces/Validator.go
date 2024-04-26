// Code generated by mockery v2.42.1. DO NOT EDIT.

package mocks

import mock "github.com/stretchr/testify/mock"

// Validator is an autogenerated mock type for the Validator type
type Validator struct {
	mock.Mock
}

// ValidateDocument provides a mock function with given fields: reference, document
func (_m *Validator) ValidateDocument(reference string, document interface{}) error {
	ret := _m.Called(reference, document)

	if len(ret) == 0 {
		panic("no return value specified for ValidateDocument")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(string, interface{}) error); ok {
		r0 = rf(reference, document)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// ValidateStruct provides a mock function with given fields: data
func (_m *Validator) ValidateStruct(data interface{}) error {
	ret := _m.Called(data)

	if len(ret) == 0 {
		panic("no return value specified for ValidateStruct")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(interface{}) error); ok {
		r0 = rf(data)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NewValidator creates a new instance of Validator. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewValidator(t interface {
	mock.TestingT
	Cleanup(func())
}) *Validator {
	mock := &Validator{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
