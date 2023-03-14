package mocks

import (
	"bytes"
	"net/http"

	"github.com/newrelic/newrelic-diagnostics-cli/helpers/httpHelper"
	"github.com/stretchr/testify/mock"
)

type MAttachDeps struct {
	mock.Mock
}

func (m MAttachDeps) GetFileSize(file string) int64 {
	ret := m.Called(file)

	var r0 int64
	if ret.Get(0) != nil {
		r0 = ret.Get(0).(int64)
	}

	return r0
}

func (m MAttachDeps) GetReader(file string) (*bytes.Reader, error) {
	ret := m.Called(file)

	var r0 *bytes.Reader
	if ret.Get(0) != nil {
		r0 = ret.Get(0).(*bytes.Reader)
	}

	var r1 error
	if ret.Get(1) != nil {
		r1 = ret.Get(1).(error)
	}

	return r0, r1
}

func (m MAttachDeps) GetWrapper(file *bytes.Reader, fileSize int64, filename string, attachmentKey string) httpHelper.RequestWrapper {
	ret := m.Called(file, fileSize, filename, attachmentKey)

	var r0 httpHelper.RequestWrapper
	if ret.Get(0) != nil {
		r0 = ret.Get(0).(httpHelper.RequestWrapper)
	}

	return r0
}

func (m MAttachDeps) GetUrlsToReturn(res *http.Response) (*string, error) {
	ret := m.Called(res)

	var r0 *string
	if ret.Get(0) != nil {
		r0 = ret.Get(0).(*string)
	}

	var r1 error
	if ret.Get(1) != nil {
		r1 = ret.Get(1).(error)
	}

	return r0, r1
}
