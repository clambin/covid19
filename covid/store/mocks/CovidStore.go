// Code generated by mockery v2.12.1. DO NOT EDIT.

package mocks

import (
	models "github.com/clambin/covid19/models"
	mock "github.com/stretchr/testify/mock"

	testing "testing"

	time "time"
)

// CovidStore is an autogenerated mock type for the CovidStore type
type CovidStore struct {
	mock.Mock
}

// Add provides a mock function with given fields: entries
func (_m *CovidStore) Add(entries []models.CountryEntry) error {
	ret := _m.Called(entries)

	var r0 error
	if rf, ok := ret.Get(0).(func([]models.CountryEntry) error); ok {
		r0 = rf(entries)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// CountEntriesByTime provides a mock function with given fields: from, to
func (_m *CovidStore) CountEntriesByTime(from time.Time, to time.Time) (map[time.Time]int, error) {
	ret := _m.Called(from, to)

	var r0 map[time.Time]int
	if rf, ok := ret.Get(0).(func(time.Time, time.Time) map[time.Time]int); ok {
		r0 = rf(from, to)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(map[time.Time]int)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(time.Time, time.Time) error); ok {
		r1 = rf(from, to)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetAll provides a mock function with given fields:
func (_m *CovidStore) GetAll() ([]models.CountryEntry, error) {
	ret := _m.Called()

	var r0 []models.CountryEntry
	if rf, ok := ret.Get(0).(func() []models.CountryEntry); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]models.CountryEntry)
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

// GetAllCountryNames provides a mock function with given fields:
func (_m *CovidStore) GetAllCountryNames() ([]string, error) {
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

// GetAllForCountryName provides a mock function with given fields: name
func (_m *CovidStore) GetAllForCountryName(name string) ([]models.CountryEntry, error) {
	ret := _m.Called(name)

	var r0 []models.CountryEntry
	if rf, ok := ret.Get(0).(func(string) []models.CountryEntry); ok {
		r0 = rf(name)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]models.CountryEntry)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(name)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetAllForRange provides a mock function with given fields: from, to
func (_m *CovidStore) GetAllForRange(from time.Time, to time.Time) ([]models.CountryEntry, error) {
	ret := _m.Called(from, to)

	var r0 []models.CountryEntry
	if rf, ok := ret.Get(0).(func(time.Time, time.Time) []models.CountryEntry); ok {
		r0 = rf(from, to)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]models.CountryEntry)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(time.Time, time.Time) error); ok {
		r1 = rf(from, to)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetFirstEntry provides a mock function with given fields:
func (_m *CovidStore) GetFirstEntry() (time.Time, bool, error) {
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

// GetLatestForCountries provides a mock function with given fields: countryNames
func (_m *CovidStore) GetLatestForCountries(countryNames []string) (map[string]models.CountryEntry, error) {
	ret := _m.Called(countryNames)

	var r0 map[string]models.CountryEntry
	if rf, ok := ret.Get(0).(func([]string) map[string]models.CountryEntry); ok {
		r0 = rf(countryNames)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(map[string]models.CountryEntry)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func([]string) error); ok {
		r1 = rf(countryNames)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetLatestForCountriesByTime provides a mock function with given fields: countryNames, endTime
func (_m *CovidStore) GetLatestForCountriesByTime(countryNames []string, endTime time.Time) (map[string]models.CountryEntry, error) {
	ret := _m.Called(countryNames, endTime)

	var r0 map[string]models.CountryEntry
	if rf, ok := ret.Get(0).(func([]string, time.Time) map[string]models.CountryEntry); ok {
		r0 = rf(countryNames, endTime)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(map[string]models.CountryEntry)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func([]string, time.Time) error); ok {
		r1 = rf(countryNames, endTime)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetTotalsPerDay provides a mock function with given fields:
func (_m *CovidStore) GetTotalsPerDay() ([]models.CountryEntry, error) {
	ret := _m.Called()

	var r0 []models.CountryEntry
	if rf, ok := ret.Get(0).(func() []models.CountryEntry); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]models.CountryEntry)
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

// NewCovidStore creates a new instance of CovidStore. It also registers the testing.TB interface on the mock and a cleanup function to assert the mocks expectations.
func NewCovidStore(t testing.TB) *CovidStore {
	mock := &CovidStore{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
