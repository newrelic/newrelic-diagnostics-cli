package mocks

import (
	"github.com/newrelic/newrelic-diagnostics-cli/domain/entity"
	"github.com/stretchr/testify/mock"
)

type MockDotNetTLSRegKey struct {
	mock.Mock
}

func (m *MockDotNetTLSRegKey) ValidateTLSRegKeys() (*entity.TLSRegKey, error) {
	ret := m.Called()

	var r0 *entity.TLSRegKey
	if ret.Get(0) != nil {
		r0 = ret.Get(0).(*entity.TLSRegKey)
	}

	var r1 error

	if ret.Get(1) != nil {
		r1 = ret.Get(1).(error)
	}

	return r0, r1
}

func (m *MockDotNetTLSRegKey) ValidateSchUseStrongCryptoKeys(path string) (*int, error) {
	ret := m.Called()

	var r0 *int
	if ret.Get(0) != nil {
		r0 = ret.Get(0).(*int)
	}

	var r1 error

	if ret.Get(1) != nil {
		r1 = ret.Get(1).(error)
	}

	return r0, r1
}
