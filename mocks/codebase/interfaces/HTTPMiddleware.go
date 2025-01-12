// Code generated by mockery v2.49.1. DO NOT EDIT.

package mocks

import (
	http "net/http"

	mock "github.com/stretchr/testify/mock"
)

// HTTPMiddleware is an autogenerated mock type for the HTTPMiddleware type
type HTTPMiddleware struct {
	mock.Mock
}

// HTTPBasicAuth provides a mock function with given fields: next
func (_m *HTTPMiddleware) HTTPBasicAuth(next http.Handler) http.Handler {
	ret := _m.Called(next)

	if len(ret) == 0 {
		panic("no return value specified for HTTPBasicAuth")
	}

	var r0 http.Handler
	if rf, ok := ret.Get(0).(func(http.Handler) http.Handler); ok {
		r0 = rf(next)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(http.Handler)
		}
	}

	return r0
}

// HTTPBearerAuth provides a mock function with given fields: next
func (_m *HTTPMiddleware) HTTPBearerAuth(next http.Handler) http.Handler {
	ret := _m.Called(next)

	if len(ret) == 0 {
		panic("no return value specified for HTTPBearerAuth")
	}

	var r0 http.Handler
	if rf, ok := ret.Get(0).(func(http.Handler) http.Handler); ok {
		r0 = rf(next)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(http.Handler)
		}
	}

	return r0
}

// HTTPCache provides a mock function with given fields: next
func (_m *HTTPMiddleware) HTTPCache(next http.Handler) http.Handler {
	ret := _m.Called(next)

	if len(ret) == 0 {
		panic("no return value specified for HTTPCache")
	}

	var r0 http.Handler
	if rf, ok := ret.Get(0).(func(http.Handler) http.Handler); ok {
		r0 = rf(next)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(http.Handler)
		}
	}

	return r0
}

// HTTPMultipleAuth provides a mock function with given fields: next
func (_m *HTTPMiddleware) HTTPMultipleAuth(next http.Handler) http.Handler {
	ret := _m.Called(next)

	if len(ret) == 0 {
		panic("no return value specified for HTTPMultipleAuth")
	}

	var r0 http.Handler
	if rf, ok := ret.Get(0).(func(http.Handler) http.Handler); ok {
		r0 = rf(next)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(http.Handler)
		}
	}

	return r0
}

// HTTPPermissionACL provides a mock function with given fields: permissionCode
func (_m *HTTPMiddleware) HTTPPermissionACL(permissionCode string) func(http.Handler) http.Handler {
	ret := _m.Called(permissionCode)

	if len(ret) == 0 {
		panic("no return value specified for HTTPPermissionACL")
	}

	var r0 func(http.Handler) http.Handler
	if rf, ok := ret.Get(0).(func(string) func(http.Handler) http.Handler); ok {
		r0 = rf(permissionCode)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(func(http.Handler) http.Handler)
		}
	}

	return r0
}

// NewHTTPMiddleware creates a new instance of HTTPMiddleware. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewHTTPMiddleware(t interface {
	mock.TestingT
	Cleanup(func())
}) *HTTPMiddleware {
	mock := &HTTPMiddleware{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
