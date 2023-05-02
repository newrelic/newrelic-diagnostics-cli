package mocks

import (
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
	"github.com/stretchr/testify/mock"
)

type MPipVersionDeps struct {
	mock.Mock
}

func (m MPipVersionDeps) CheckPipVersion(pipCmd string) tasks.Result {
	ret := m.Called(pipCmd)

	var r0 tasks.Result
	if ret.Get(0) != nil {
		r0 = ret.Get(0).(tasks.Result)
	}

	return r0
}
