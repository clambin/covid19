// Code generated by mockery v0.0.0-dev. DO NOT EDIT.

package mocks

import (
	coviddb "github.com/clambin/covid19/coviddb"
	mock "github.com/stretchr/testify/mock"

	time "time"
)

// DB is an autogenerated mock type for the DB type
type DB struct {
	mock.Mock
}

// Add provides a mock function with given fields: _a0
func (_m *DB) Add(_a0 []coviddb.CountryEntry) error {
	ret := _m.Called(_a0)

	var r0 error
	if rf, ok := ret.Get(0).(func([]coviddb.CountryEntry) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetAllCountryCodes provides a mock function with given fields:
func (_m *DB) GetAllCountryCodes() ([]string, error) {
	ret := _m.Called()

	var r0 []string
	if rf, ok := ret.Get(0).(func() []string); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]string)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetFirstEntry provides a mock function with given fields:
func (_m *DB) GetFirstEntry() (time.Time, bool, error) {
	ret := _m.Called()

	var r0 time.Time
	if rf, ok := ret.Get(0).(func() time.Time); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(time.Time)
	}

	var r1 bool
	if rf, ok := ret.Get(1).(func() bool); ok {
		r1 = rf()
	} else {
		r1 = ret.Get(1).(bool)
	}

	var r2 error
	if rf, ok := ret.Get(2).(func() error); ok {
		r2 = rf()
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// GetLastForCountry provides a mock function with given fields: _a0
func (_m *DB) GetLastForCountry(_a0 string) (*coviddb.CountryEntry, bool, error) {
	ret := _m.Called(_a0)

	var r0 *coviddb.CountryEntry
	if rf, ok := ret.Get(0).(func(string) *coviddb.CountryEntry); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*coviddb.CountryEntry)
		}
	}

	var r1 bool
	if rf, ok := ret.Get(1).(func(string) bool); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Get(1).(bool)
	}

	var r2 error
	if rf, ok := ret.Get(2).(func(string) error); ok {
		r2 = rf(_a0)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// List provides a mock function with given fields:
func (_m *DB) List() ([]coviddb.CountryEntry, error) {
	ret := _m.Called()

	var r0 []coviddb.CountryEntry
	if rf, ok := ret.Get(0).(func() []coviddb.CountryEntry); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]coviddb.CountryEntry)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListLatestByCountry provides a mock function with given fields:
func (_m *DB) ListLatestByCountry() (map[string]time.Time, error) {
	ret := _m.Called()

	var r0 map[string]time.Time
	if rf, ok := ret.Get(0).(func() map[string]time.Time); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(map[string]time.Time)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
