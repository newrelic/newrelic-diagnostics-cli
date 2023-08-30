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
description: |-
  test
  test
  test
type: bash
os: darwin
outputFiles: 
  - 'file*.log'
  - '*file.log'
`)
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

type MockScriptRunner struct{}

func (m *MockScriptRunner) GetUUID() string {
	return "1234-1234-1234-1234"
}

func (m *MockScriptRunner) ContinueIfExists(savepath string) bool {
	return true
}

func (m *MockScriptRunner) SaveToDisk(body []byte, savepath string) error {
	return nil
}
func (m *MockScriptRunner) RunScript(body []byte, savepath string, scriptOptions string) ([]byte, error) {
	return []byte("mock"), nil
}
