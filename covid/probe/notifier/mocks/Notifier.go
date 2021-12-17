// Code generated by mockery v0.0.0-dev. DO NOT EDIT.

package mocks

import (
	models "github.com/clambin/covid19/models"
	mock "github.com/stretchr/testify/mock"
)

// Notifier is an autogenerated mock type for the Notifier type
type Notifier struct {
	mock.Mock
}

// Notify provides a mock function with given fields: entries
func (_m *Notifier) Notify(entries []models.CountryEntry) error {
	ret := _m.Called(entries)

	var r0 error
	if rf, ok := ret.Get(0).(func([]models.CountryEntry) error); ok {
		r0 = rf(entries)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
