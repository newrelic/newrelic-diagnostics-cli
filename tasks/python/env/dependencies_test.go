package env

import (
	"reflect"
	"testing"

	"github.com/newrelic/newrelic-diagnostics-cli/domain/repository"
	"github.com/newrelic/newrelic-diagnostics-cli/mocks"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
)

var _ = Describe("Python/Env/Dependencies", func() {
	var p PythonEnvDependencies

	Describe("Identifier()", func() {
		It("Should return correct identifier", func() {
			expectedIdentifier := tasks.Identifier{
				Category:    "Python",
				Subcategory: "Env",
				Name:        "Dependencies",
			}

			Expect(p.Identifier()).To(Equal(expectedIdentifier))
		})

	})

	Describe("Explain()", func() {
		It("Should return correct task explanations string", func() {
			expectedExplanation := "Collect Python application packages"
			Expect(p.Explain()).To(Equal(expectedExplanation))
		})
	})

	Describe("Dependencies()", func() {
		It("Should return an expected slice of dependencies", func() {
			expectedDependencies := []string{"Python/Config/Agent"}
			Expect(p.Dependencies()).To(Equal(expectedDependencies))
		})
	})

	Describe("Execute()", func() {
		var (
			result   tasks.Result
			options  tasks.Options
			upstream map[string]tasks.Result
		)

		JustBeforeEach(func() {
			result = p.Execute(options, upstream)
		})

		Context("upstream dependency task failed", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Python/Config/Agent": {
						Status: tasks.Failure,
					},
				}
			})

			It("should return an expected none result status", func() {
				Expect(result.Status).To(Equal(tasks.None))
			})

		})

	})
})

func TestPythonEnvDependencies_getProjectDependencies(t *testing.T) {
	mPipEnv := new(mocks.MPipVersionDeps)
	stream := make(chan string)

	filesToCopy := []tasks.FileCopyEnvelope{
		{Path: "pipFreeze.txt", Stream: stream},
	}
	type fields struct {
		iPipEnvVersion repository.IPipEnvVersion
	}
	tests := []struct {
		name           string
		fields         fields
		want           tasks.Result
		mockPipReturn  tasks.Result
		mockPip3Return tasks.Result
	}{
		// TODO: Add test cases.
		{
			name: "Both Success",
			fields: fields{
				iPipEnvVersion: mPipEnv,
			},
			want: tasks.Result{
				Status:      tasks.Success,
				Summary:     "SUCCESS\nSUCCESS",
				Payload:     []string{"PAYLOAD"},
				FilesToCopy: []tasks.FileCopyEnvelope{filesToCopy[0], filesToCopy[0]},
			},
			mockPipReturn: tasks.Result{
				Status:      tasks.Info,
				Summary:     "SUCCESS",
				Payload:     []string{"PAYLOAD"},
				FilesToCopy: filesToCopy,
			},
			mockPip3Return: tasks.Result{
				Status:      tasks.Info,
				Summary:     "SUCCESS",
				Payload:     []string{"PAYLOAD"},
				FilesToCopy: filesToCopy,
			},
		},
		{
			name: "pip freeze returns an error and pip3 freeze returns a success",
			fields: fields{
				iPipEnvVersion: mPipEnv,
			},
			want: tasks.Result{
				Status:      tasks.Warning,
				Summary:     "FAILURE\nSUCCESS",
				Payload:     []string{"PAYLOAD"},
				FilesToCopy: []tasks.FileCopyEnvelope{filesToCopy[0]},
			},
			mockPipReturn: tasks.Result{
				Status:  tasks.Error,
				Summary: "FAILURE",
			},
			mockPip3Return: tasks.Result{
				Status:      tasks.Info,
				Summary:     "SUCCESS",
				Payload:     []string{"PAYLOAD"},
				FilesToCopy: filesToCopy,
			},
		},
		{
			name: "pip3 freeze returns an error and pip freeze returns a success",
			fields: fields{
				iPipEnvVersion: mPipEnv,
			},
			want: tasks.Result{
				Status:      tasks.Warning,
				Summary:     "FAILURE\nSUCCESS",
				Payload:     []string{"PAYLOAD"},
				FilesToCopy: []tasks.FileCopyEnvelope{filesToCopy[0]},
			},
			mockPipReturn: tasks.Result{
				Status:      tasks.Info,
				Summary:     "SUCCESS",
				Payload:     []string{"PAYLOAD"},
				FilesToCopy: filesToCopy,
			},
			mockPip3Return: tasks.Result{
				Status:  tasks.Error,
				Summary: "FAILURE",
			},
		},
		{
			name: "Both return errors",
			fields: fields{
				iPipEnvVersion: mPipEnv,
			},
			want: tasks.Result{
				Status:  tasks.Error,
				Summary: "FAILURE\nFAILURE",
			},
			mockPipReturn: tasks.Result{
				Status:  tasks.Error,
				Summary: "FAILURE",
			},
			mockPip3Return: tasks.Result{
				Status:  tasks.Error,
				Summary: "FAILURE",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mPipEnv.On("CheckPipVersion", mock.Anything).Return(tt.mockPipReturn).Once()
			mPipEnv.On("CheckPipVersion", mock.Anything).Return(tt.mockPip3Return).Once()
			tr := PythonEnvDependencies{
				iPipEnvVersion: tt.fields.iPipEnvVersion,
			}
			if got := tr.getProjectDependencies(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PythonEnvDependencies.getProjectDependencies() = %v, want %v", got, tt.want)
			}
		})
	}
}
