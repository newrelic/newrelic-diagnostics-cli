package mocks

import (
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
	"github.com/stretchr/testify/mock"
)

type MPythonVersionDeps struct {
	mock.Mock
}

func (m MPythonVersionDeps) CheckPythonVersion(pythonCmd string) tasks.Result {
	ret := m.Called(pythonCmd)

	var r0 tasks.Result
	if ret.Get(0) != nil {
		r0 = ret.Get(0).(tasks.Result)
	}

	return r0
}
