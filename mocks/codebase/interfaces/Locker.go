// Code generated by mockery v2.42.1. DO NOT EDIT.

package mocks

import (
	context "context"

	mock "github.com/stretchr/testify/mock"

	time "time"
)

// Locker is an autogenerated mock type for the Locker type
type Locker struct {
	mock.Mock
}

// Disconnect provides a mock function with given fields: ctx
func (_m *Locker) Disconnect(ctx context.Context) error {
	ret := _m.Called(ctx)

	if len(ret) == 0 {
		panic("no return value specified for Disconnect")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context) error); ok {
		r0 = rf(ctx)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// HasBeenLocked provides a mock function with given fields: key
func (_m *Locker) HasBeenLocked(key string) bool {
	ret := _m.Called(key)

	if len(ret) == 0 {
		panic("no return value specified for HasBeenLocked")
	}

	var r0 bool
	if rf, ok := ret.Get(0).(func(string) bool); ok {
		r0 = rf(key)
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// IsLocked provides a mock function with given fields: key
func (_m *Locker) IsLocked(key string) bool {
	ret := _m.Called(key)

	if len(ret) == 0 {
		panic("no return value specified for IsLocked")
	}

	var r0 bool
	if rf, ok := ret.Get(0).(func(string) bool); ok {
		r0 = rf(key)
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// Lock provides a mock function with given fields: key, timeout
func (_m *Locker) Lock(key string, timeout time.Duration) (func(), error) {
	ret := _m.Called(key, timeout)

	if len(ret) == 0 {
		panic("no return value specified for Lock")
	}

	var r0 func()
	var r1 error
	if rf, ok := ret.Get(0).(func(string, time.Duration) (func(), error)); ok {
		return rf(key, timeout)
	}
	if rf, ok := ret.Get(0).(func(string, time.Duration) func()); ok {
		r0 = rf(key, timeout)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(func())
		}
	}

	if rf, ok := ret.Get(1).(func(string, time.Duration) error); ok {
		r1 = rf(key, timeout)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Reset provides a mock function with given fields: key
func (_m *Locker) Reset(key string) {
	_m.Called(key)
}

// Unlock provides a mock function with given fields: key
func (_m *Locker) Unlock(key string) {
	_m.Called(key)
}

// NewLocker creates a new instance of Locker. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewLocker(t interface {
	mock.TestingT
	Cleanup(func())
}) *Locker {
	mock := &Locker{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
