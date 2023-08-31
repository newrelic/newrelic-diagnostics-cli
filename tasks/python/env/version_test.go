package env

import (
	"errors"
	"reflect"
	"testing"

	"github.com/newrelic/newrelic-diagnostics-cli/domain/repository"
	"github.com/newrelic/newrelic-diagnostics-cli/mocks"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"

	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

func TestEnv(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Env Suite")
}

func mockCommandExecuteError(name string, arg ...string) ([]byte, error) {
	return []byte{}, errors.New("mock error")
}

func mockCommandExecuteSuccess(name string, arg ...string) ([]byte, error) {
	return []byte("Python 123"), nil
}

func mockCommandExecuteBadOutput(name string, arg ...string) ([]byte, error) {
	return []byte("Not Python 123"), nil
}

var _ = Describe("PythonEnvVersion", func() {
	p := PythonEnvVersion{}

	Describe("Identifier()", func() {
		It("Should return identifier", func() {
			Expect(p.Identifier()).To(Equal(tasks.Identifier{Name: "Version", Category: "Python", Subcategory: "Env"}))
		})
	})

	Describe("Explain()", func() {
		It("should explain what the task does", func() {
			Expect(p.Explain()).To(Equal("Determine Python version"))
		})
	})

	Describe("Dependencies()", func() {
		It("should return list of dependencies' ", func() {
			Expect(p.Dependencies()).To(Equal([]string{"Python/Config/Agent"}))
		})
	})

	Describe("Execute()", func() {
		Context("When the tasks dependencies were unsuccessful", func() {
			It("should return a 'none' result.", func() {
				failedDependency := map[string]tasks.Result{
					"Python/Config/Agent": {Status: tasks.Failure},
				}
				Expect(p.Execute(tasks.Options{}, failedDependency).Status).To(Equal(tasks.None))
			})
		})

	})

	// Describe("parsePythonVersion()", func() {
	// 	It("should return ('123', true) when passed 'Python 123 '", func() {
	// 		versionString, isValid := parsePythonVersion([]byte("Python 123 "))
	// 		Expect(versionString).To(Equal("123"))
	// 		Expect(isValid).To(BeTrue())
	// 	})
	// 	It("should return ('', false) when passed 'Pyrhon 123 '", func() {
	// 		versionString, isValid := parsePythonVersion([]byte("Pyrhon 123 "))
	// 		Expect(versionString).To(Equal(""))
	// 		Expect(isValid).To(BeFalse())
	// 	})
	// })
})

func TestPythonEnvVersion_RunPythonCommands(t *testing.T) {
	mPythonEnv := new(mocks.MPythonVersionDeps)
	type fields struct {
		iPythonEnvVersion repository.IPythonEnvVersion
	}
	tests := []struct {
		name              string
		fields            fields
		want              tasks.Result
		mockPythonReturn  tasks.Result
		mockPython3Return tasks.Result
	}{
		// TODO: Add test cases.
		{
			name: "Initial Success",
			fields: fields{
				iPythonEnvVersion: mPythonEnv,
			},
			want: tasks.Result{
				Status:  tasks.Success,
				Summary: "SUCCESS\nSUCCESS",
				Payload: []string{"PAYLOAD", "PAYLOAD"},
			},
			mockPythonReturn: tasks.Result{
				Status:  tasks.Info,
				Summary: "SUCCESS",
				Payload: "PAYLOAD",
			},
			mockPython3Return: tasks.Result{
				Status:  tasks.Info,
				Summary: "SUCCESS",
				Payload: "PAYLOAD",
			},
		},
		{
			name: "python --version fails",
			fields: fields{
				iPythonEnvVersion: mPythonEnv,
			},
			want: tasks.Result{
				Status:  tasks.Warning,
				Summary: "FAIL\nSUCCESS",
				URL:     "https://docs.newrelic.com/docs/agents/python-agent/getting-started/compatibility-requirements-python-agent#basic",
				Payload: []string{"PAYLOAD"},
			},
			mockPythonReturn: tasks.Result{
				Status:  tasks.Error,
				Summary: "FAIL",
				URL:     "https://docs.newrelic.com/docs/agents/python-agent/getting-started/compatibility-requirements-python-agent#basic",
			},
			mockPython3Return: tasks.Result{
				Status:  tasks.Info,
				Summary: "SUCCESS",
				Payload: "PAYLOAD",
			},
		},
		{
			name: "python3 --version fails",
			fields: fields{
				iPythonEnvVersion: mPythonEnv,
			},
			want: tasks.Result{
				Status:  tasks.Warning,
				Summary: "FAIL\nSUCCESS",
				URL:     "https://docs.newrelic.com/docs/agents/python-agent/getting-started/compatibility-requirements-python-agent#basic",
				Payload: []string{"PAYLOAD"},
			},
			mockPythonReturn: tasks.Result{
				Status:  tasks.Info,
				Summary: "SUCCESS",
				Payload: "PAYLOAD",
			},
			mockPython3Return: tasks.Result{
				Status:  tasks.Error,
				Summary: "FAIL",
				URL:     "https://docs.newrelic.com/docs/agents/python-agent/getting-started/compatibility-requirements-python-agent#basic",
			},
		},
		{
			name: "Both python --version and python3 --version fails",
			fields: fields{
				iPythonEnvVersion: mPythonEnv,
			},
			want: tasks.Result{
				Status:  tasks.Error,
				Summary: "FAIL\nFAIL",
				URL:     "https://docs.newrelic.com/docs/agents/python-agent/getting-started/compatibility-requirements-python-agent#basic",
			},
			mockPythonReturn: tasks.Result{
				Status:  tasks.Error,
				Summary: "FAIL",
				URL:     "https://docs.newrelic.com/docs/agents/python-agent/getting-started/compatibility-requirements-python-agent#basic",
			},
			mockPython3Return: tasks.Result{
				Status:  tasks.Error,
				Summary: "FAIL",
				URL:     "https://docs.newrelic.com/docs/agents/python-agent/getting-started/compatibility-requirements-python-agent#basic",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mPythonEnv.On("CheckPythonVersion", mock.Anything).Return(tt.mockPythonReturn).Once()
			mPythonEnv.On("CheckPythonVersion", mock.Anything).Return(tt.mockPython3Return).Once()
			p := PythonEnvVersion{
				iPythonEnvVersion: tt.fields.iPythonEnvVersion,
			}
			if got := p.RunPythonCommands(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PythonEnvVersion.RunPythonCommands() = %v, want %v", got, tt.want)
			}
		})
	}
}
