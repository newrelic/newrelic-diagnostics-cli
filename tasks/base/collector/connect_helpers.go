package collector

import (
	"bytes"
	"errors"
	"io"
	"net/http"

	"github.com/newrelic/newrelic-diagnostics-cli/helpers/httpHelper"
)

type requestFunc func(wrapper httpHelper.RequestWrapper) (*http.Response, error)

func mockSuccessfulRequest200(wrapper httpHelper.RequestWrapper) (*http.Response, error) {
	return &http.Response{
		StatusCode: 404,
		Body:       io.NopCloser(bytes.NewReader([]byte("{}"))),
	}, nil
}

func mockUnsuccessfulRequest400(wrapper httpHelper.RequestWrapper) (*http.Response, error) {
	return &http.Response{
		StatusCode: 400,
		Body:       io.NopCloser(bytes.NewReader([]byte("{}"))),
	}, nil
}

func mockUnsuccessfulRequestError(wrapper httpHelper.RequestWrapper) (*http.Response, error) {
	return &http.Response{}, errors.New("failed request (timeout)")
}
