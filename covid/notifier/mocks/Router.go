// Code generated by mockery v2.15.0. DO NOT EDIT.

package mocks

import mock "github.com/stretchr/testify/mock"

// Router is an autogenerated mock type for the Router type
type Router struct {
	mock.Mock
}

// Send provides a mock function with given fields: title, message
func (_m *Router) Send(title string, message string) error {
	ret := _m.Called(title, message)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string) error); ok {
		r0 = rf(title, message)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

type mockConstructorTestingTNewRouter interface {
	mock.TestingT
	Cleanup(func())
}

// NewRouter creates a new instance of Router. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewRouter(t mockConstructorTestingTNewRouter) *Router {
	mock := &Router{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
