package mocks

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

type MockCatalogDependenciesSuccessful struct{}

func (m *MockCatalogDependenciesSuccessful) MakeRequest(name ...string) (*http.Response, error) {
	if name == nil {
		return &http.Response{
			Body: io.NopCloser(bytes.NewReader([]byte(`[{"name": "test.yml"}]`))),
		}, nil
	}
	content := []byte(`name: test
filename: test.sh
description: test
type: bash
os: darwin`)
	return &http.Response{
		Body: io.NopCloser(bytes.NewReader([]byte(`{"Content": "` + base64.StdEncoding.EncodeToString(content) + `"}`))),
	}, nil
}

type MockCatalogDependenciesErrorList struct{}

func (m *MockCatalogDependenciesErrorList) MakeRequest(name ...string) (*http.Response, error) {
	return nil, errors.New("test error")
}

type MockCatalogDependenciesErrorFile struct{}

func (m *MockCatalogDependenciesErrorFile) MakeRequest(name ...string) (*http.Response, error) {
	if name == nil {
		return &http.Response{
			Body: io.NopCloser(bytes.NewReader([]byte(`[{"name": "test.yml"}]`))),
		}, nil
	}
	return nil, errors.New("test error")
}

type MockCatalogDependencies403Error struct{}

func (m *MockCatalogDependencies403Error) MakeRequest(name ...string) (*http.Response, error) {
	header := http.Header{}
	header.Add("X-Ratelimit-Reset", fmt.Sprint(time.Now().Unix()))
	return &http.Response{
		StatusCode: 403,
		Header:     header,
		Body:       io.NopCloser(bytes.NewReader([]byte(``))),
	}, nil

}
