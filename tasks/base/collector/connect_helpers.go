package collector

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"

	"github.com/newrelic/NrDiag/helpers/httpHelper"
)

type requestFunc func(wrapper httpHelper.RequestWrapper) (*http.Response, error)

func mockSuccessfulRequest200(wrapper httpHelper.RequestWrapper) (*http.Response, error) {
	return &http.Response{
		StatusCode: 200,
		Body:       ioutil.NopCloser(bytes.NewReader([]byte("test body"))),
	}, nil
}

func mockUnsuccessfulRequest400(wrapper httpHelper.RequestWrapper) (*http.Response, error) {
	return &http.Response{
		StatusCode: 400,
		Body:       ioutil.NopCloser(bytes.NewReader([]byte("test body"))),
	}, nil
}

func mockUnsuccessfulRequestError(wrapper httpHelper.RequestWrapper) (*http.Response, error) {
	return &http.Response{}, errors.New("Failed request (timeout)")
}
