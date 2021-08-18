package probe_test

import (
	"context"
	covidDBMock "github.com/clambin/covid19/coviddb/mocks"
	popDBMock "github.com/clambin/covid19/population/db/mocks"
	"github.com/clambin/covid19/population/probe"
	probeMock "github.com/clambin/covid19/population/probe/mocks"
	"github.com/stretchr/testify/mock"
	"testing"
)

func TestProbe_Update(t *testing.T) {
	covidDB := &covidDBMock.DB{}
	popDB := &popDBMock.DB{}
	apiClient := &probeMock.APIClient{}

	p := probe.Create("1234", popDB, covidDB)
	p.APIClient = apiClient

	covidDB.On("GetAllCountryCodes").Return([]string{"BE", "US"}, nil)
	apiClient.On("GetPopulation", mock.Anything, "United States").Return(int64(330), nil)
	apiClient.On("GetPopulation", mock.Anything, "Belgium").Return(int64(11), nil)
	popDB.On("Add", "US", int64(330)).Return(nil)
	popDB.On("Add", "BE", int64(11)).Return(nil)

	p.Update(context.TODO())

	mock.AssertExpectationsForObjects(t, covidDB, popDB, apiClient)
}
