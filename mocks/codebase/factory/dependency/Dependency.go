// Code generated by mockery v2.42.1. DO NOT EDIT.

package mocks

import (
	interfaces "github.com/golangid/candi/codebase/interfaces"
	mock "github.com/stretchr/testify/mock"

	types "github.com/golangid/candi/codebase/factory/types"
)

// Dependency is an autogenerated mock type for the Dependency type
type Dependency struct {
	mock.Mock
}

// AddBroker provides a mock function with given fields: brokerType, b
func (_m *Dependency) AddBroker(brokerType types.Worker, b interfaces.Broker) {
	_m.Called(brokerType, b)
}

// AddExtended provides a mock function with given fields: key, value
func (_m *Dependency) AddExtended(key string, value interface{}) {
	_m.Called(key, value)
}

// FetchBroker provides a mock function with given fields: _a0
func (_m *Dependency) FetchBroker(_a0 func(types.Worker, interfaces.Broker)) {
	_m.Called(_a0)
}

// GetBroker provides a mock function with given fields: _a0
func (_m *Dependency) GetBroker(_a0 types.Worker) interfaces.Broker {
	ret := _m.Called(_a0)

	if len(ret) == 0 {
		panic("no return value specified for GetBroker")
	}

	var r0 interfaces.Broker
	if rf, ok := ret.Get(0).(func(types.Worker) interfaces.Broker); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(interfaces.Broker)
		}
	}

	return r0
}

// GetExtended provides a mock function with given fields: key
func (_m *Dependency) GetExtended(key string) interface{} {
	ret := _m.Called(key)

	if len(ret) == 0 {
		panic("no return value specified for GetExtended")
	}

	var r0 interface{}
	if rf, ok := ret.Get(0).(func(string) interface{}); ok {
		r0 = rf(key)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(interface{})
		}
	}

	return r0
}

// GetKey provides a mock function with given fields:
func (_m *Dependency) GetKey() interfaces.RSAKey {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for GetKey")
	}

	var r0 interfaces.RSAKey
	if rf, ok := ret.Get(0).(func() interfaces.RSAKey); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(interfaces.RSAKey)
		}
	}

	return r0
}

// GetLocker provides a mock function with given fields:
func (_m *Dependency) GetLocker() interfaces.Locker {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for GetLocker")
	}

	var r0 interfaces.Locker
	if rf, ok := ret.Get(0).(func() interfaces.Locker); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(interfaces.Locker)
		}
	}

	return r0
}

// GetMiddleware provides a mock function with given fields:
func (_m *Dependency) GetMiddleware() interfaces.Middleware {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for GetMiddleware")
	}

	var r0 interfaces.Middleware
	if rf, ok := ret.Get(0).(func() interfaces.Middleware); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(interfaces.Middleware)
		}
	}

	return r0
}

// GetMongoDatabase provides a mock function with given fields:
func (_m *Dependency) GetMongoDatabase() interfaces.MongoDatabase {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for GetMongoDatabase")
	}

	var r0 interfaces.MongoDatabase
	if rf, ok := ret.Get(0).(func() interfaces.MongoDatabase); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(interfaces.MongoDatabase)
		}
	}

	return r0
}

// GetRedisPool provides a mock function with given fields:
func (_m *Dependency) GetRedisPool() interfaces.RedisPool {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for GetRedisPool")
	}

	var r0 interfaces.RedisPool
	if rf, ok := ret.Get(0).(func() interfaces.RedisPool); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(interfaces.RedisPool)
		}
	}

	return r0
}

// GetSQLDatabase provides a mock function with given fields:
func (_m *Dependency) GetSQLDatabase() interfaces.SQLDatabase {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for GetSQLDatabase")
	}

	var r0 interfaces.SQLDatabase
	if rf, ok := ret.Get(0).(func() interfaces.SQLDatabase); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(interfaces.SQLDatabase)
		}
	}

	return r0
}

// GetValidator provides a mock function with given fields:
func (_m *Dependency) GetValidator() interfaces.Validator {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for GetValidator")
	}

	var r0 interfaces.Validator
	if rf, ok := ret.Get(0).(func() interfaces.Validator); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(interfaces.Validator)
		}
	}

	return r0
}

// SetKey provides a mock function with given fields: i
func (_m *Dependency) SetKey(i interfaces.RSAKey) {
	_m.Called(i)
}

// SetLocker provides a mock function with given fields: v
func (_m *Dependency) SetLocker(v interfaces.Locker) {
	_m.Called(v)
}

// SetMiddleware provides a mock function with given fields: mw
func (_m *Dependency) SetMiddleware(mw interfaces.Middleware) {
	_m.Called(mw)
}

// SetValidator provides a mock function with given fields: v
func (_m *Dependency) SetValidator(v interfaces.Validator) {
	_m.Called(v)
}

// NewDependency creates a new instance of Dependency. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewDependency(t interface {
	mock.TestingT
	Cleanup(func())
}) *Dependency {
	mock := &Dependency{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
