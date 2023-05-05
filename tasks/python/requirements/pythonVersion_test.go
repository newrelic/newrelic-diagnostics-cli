package requirements

import (
	"reflect"
	"testing"

	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Python/Requirements/PythonVersion", func() {
	var p PythonRequirementsPythonVersion //instance of our task struct to be used in tests

	Describe("Identifier()", func() {
		It("Should return correct identifier", func() {
			expectedIdentifier := tasks.Identifier{
				Category:    "Python",
				Subcategory: "Requirements",
				Name:        "PythonVersion",
			}

			Expect(p.Identifier()).To(Equal(expectedIdentifier))
		})
	})

	Describe("Explain()", func() {
		It("Should return correct task explanations string", func() {
			expectedExplanation := "Check Python version compatibility with New Relic Python agent"
			Expect(p.Explain()).To(Equal(expectedExplanation))
		})
	})

	Describe("Dependencies()", func() {
		It("Should return an expected slice of dependencies", func() {
			expectedDependencies := []string{"Python/Env/Version",
				"Python/Agent/Version"}
			Expect(p.Dependencies()).To(Equal(expectedDependencies))
		})
	})
})

func TestPythonRequirementsPythonVersion_Execute(t *testing.T) {
	type args struct {
		options  tasks.Options
		upstream map[string]tasks.Result
	}
	tests := []struct {
		name string
		tr   PythonRequirementsPythonVersion
		args args
		want tasks.Result
	}{
		// TODO: Add test cases.
		{
			name: "Python/Env/Version returns an error",
			tr:   PythonRequirementsPythonVersion{},
			args: args{
				options: tasks.Options{
					Options: map[string]string{"Option1": "option"},
				},
				upstream: map[string]tasks.Result{
					"Python/Env/Version": {
						Status:  tasks.Error,
						Summary: "ERROR",
					},
				},
			},
			want: tasks.Result{
				Status:  tasks.None,
				Summary: "Python version not detected. This task didn't run.",
			},
		},
		{
			name: "Python/Agent/Version returns non info",
			tr:   PythonRequirementsPythonVersion{},
			args: args{
				options: tasks.Options{
					Options: map[string]string{"Option1": "option"},
				},
				upstream: map[string]tasks.Result{
					"Python/Agent/Version": {
						Status:  tasks.Error,
						Summary: "ERROR",
					},
				},
			},
			want: tasks.Result{
				Status:  tasks.None,
				Summary: "Python Agent version not detected. This task didn't run.",
			},
		},
		{
			name: "No Python Versions are in Compatibility Map",
			tr:   PythonRequirementsPythonVersion{},
			args: args{
				options: tasks.Options{
					Options: map[string]string{"Option1": "option"},
				},
				upstream: map[string]tasks.Result{
					"Python/Env/Version": {
						Status:  tasks.Info,
						Summary: "SUMMARY",
						Payload: []string{"1.1"},
					},
					"Python/Agent/Version": {
						Status:  tasks.Info,
						Summary: "SUMMARY",
						Payload: "8.3.0",
					},
				},
			},
			want: tasks.Result{
				Status:  tasks.Failure,
				Summary: "None of your versions of Python (1.1) are supported by the Python Agent. Please review our documentation on version requirements",
				URL:     "https://docs.newrelic.com/docs/agents/python-agent/getting-started/compatibility-requirements-python-agent#basic",
			},
		},
		{
			name: "Python Agent Version comes back as not a version",
			tr:   PythonRequirementsPythonVersion{},
			args: args{
				options: tasks.Options{
					Options: map[string]string{"Option1": "option"},
				},
				upstream: map[string]tasks.Result{
					"Python/Env/Version": {
						Status:  tasks.Info,
						Summary: "SUMMARY",
						Payload: []string{"3.3"},
					},
					"Python/Agent/Version": {
						Status:  tasks.Info,
						Summary: "SUMMARY",
						Payload: "BADVERSION",
					},
				},
			},
			want: tasks.Result{
				Status:  tasks.Error,
				Summary: "We ran into an error while parsing your current agent version BADVERSION. unable to parse version: BADVERSION",
			},
		},
		{
			name: "Python Version is not compatible with agent version",
			tr:   PythonRequirementsPythonVersion{},
			args: args{
				options: tasks.Options{
					Options: map[string]string{"Option1": "option"},
				},
				upstream: map[string]tasks.Result{
					"Python/Env/Version": {
						Status:  tasks.Info,
						Summary: "SUMMARY",
						Payload: []string{"3.3"},
					},
					"Python/Agent/Version": {
						Status:  tasks.Info,
						Summary: "SUMMARY",
						Payload: "8.3.0",
					},
				},
			},
			want: tasks.Result{
				Status:  tasks.Failure,
				Summary: "None of your versions of Python (3.3) are supported by the Python Agent. Please review our documentation on version requirements",
				URL:     "https://docs.newrelic.com/docs/agents/python-agent/getting-started/compatibility-requirements-python-agent#basic",
			},
		},
		{
			name: "Some Python Version is compatible",
			tr:   PythonRequirementsPythonVersion{},
			args: args{
				options: tasks.Options{
					Options: map[string]string{"Option1": "option"},
				},
				upstream: map[string]tasks.Result{
					"Python/Env/Version": {
						Status:  tasks.Info,
						Summary: "SUMMARY",
						Payload: []string{"3.3", "3.11"},
					},
					"Python/Agent/Version": {
						Status:  tasks.Info,
						Summary: "SUMMARY",
						Payload: "8.3.1",
					},
				},
			},
			want: tasks.Result{
				Status:  tasks.Warning,
				Summary: "Your 3.3 Python version is not supported by this specific Python Agent Version (8.3.1). You'll have to use a different version of the Python Agent, 2.42.0.35 as the minimum, to ensure the agent works as expected.\nYour 3.11 Python version(s) are supported by our Python Agent",
				URL:     "https://docs.newrelic.com/docs/agents/python-agent/getting-started/compatibility-requirements-python-agent#basic",
			},
		},
		{
			name: "Some Python Version is compatible and one has no compatiblity",
			tr:   PythonRequirementsPythonVersion{},
			args: args{
				options: tasks.Options{
					Options: map[string]string{"Option1": "option"},
				},
				upstream: map[string]tasks.Result{
					"Python/Env/Version": {
						Status:  tasks.Info,
						Summary: "SUMMARY",
						Payload: []string{"3.1", "3.3", "3.11"},
					},
					"Python/Agent/Version": {
						Status:  tasks.Info,
						Summary: "SUMMARY",
						Payload: "8.3.1",
					},
				},
			},
			want: tasks.Result{
				Status:  tasks.Warning,
				Summary: "Some of your versions of Python (3.1) are not supported by the Python Agent. Please review our documentation on version requirements.\nYour 3.3 Python version is not supported by this specific Python Agent Version (8.3.1). You'll have to use a different version of the Python Agent, 2.42.0.35 as the minimum, to ensure the agent works as expected.\nYour 3.11 Python version(s) are supported by our Python Agent",
				URL:     "https://docs.newrelic.com/docs/agents/python-agent/getting-started/compatibility-requirements-python-agent#basic",
			},
		},
		{
			name: "Some Python Version is compatible and multiple have no compatiblity",
			tr:   PythonRequirementsPythonVersion{},
			args: args{
				options: tasks.Options{
					Options: map[string]string{"Option1": "option"},
				},
				upstream: map[string]tasks.Result{
					"Python/Env/Version": {
						Status:  tasks.Info,
						Summary: "SUMMARY",
						Payload: []string{"3.1", "1.1", "3.3", "3.11"},
					},
					"Python/Agent/Version": {
						Status:  tasks.Info,
						Summary: "SUMMARY",
						Payload: "8.3.1",
					},
				},
			},
			want: tasks.Result{
				Status:  tasks.Warning,
				Summary: "Some of your versions of Python (3.1,1.1) are not supported by the Python Agent. Please review our documentation on version requirements.\nYour 3.3 Python version is not supported by this specific Python Agent Version (8.3.1). You'll have to use a different version of the Python Agent, 2.42.0.35 as the minimum, to ensure the agent works as expected.\nYour 3.11 Python version(s) are supported by our Python Agent",
				URL:     "https://docs.newrelic.com/docs/agents/python-agent/getting-started/compatibility-requirements-python-agent#basic",
			},
		},
		{
			name: "Multiple Python Version is compatible and multiple have no compatiblity",
			tr:   PythonRequirementsPythonVersion{},
			args: args{
				options: tasks.Options{
					Options: map[string]string{"Option1": "option"},
				},
				upstream: map[string]tasks.Result{
					"Python/Env/Version": {
						Status:  tasks.Info,
						Summary: "SUMMARY",
						Payload: []string{"3.1", "1.1", "3.3", "3.9", "3.8"},
					},
					"Python/Agent/Version": {
						Status:  tasks.Info,
						Summary: "SUMMARY",
						Payload: "5.20.4",
					},
				},
			},
			want: tasks.Result{
				Status:  tasks.Warning,
				Summary: "Some of your versions of Python (3.1,1.1) are not supported by the Python Agent. Please review our documentation on version requirements.\nYour 3.3 Python version is not supported by this specific Python Agent Version (5.20.4). You'll have to use a different version of the Python Agent, 2.42.0.35 as the minimum, to ensure the agent works as expected.\nYour 3.9,3.8 Python version(s) are supported by our Python Agent",
				URL:     "https://docs.newrelic.com/docs/agents/python-agent/getting-started/compatibility-requirements-python-agent#basic",
			},
		},
		{
			name: "Python Version is compatible",
			tr:   PythonRequirementsPythonVersion{},
			args: args{
				options: tasks.Options{
					Options: map[string]string{"Option1": "option"},
				},
				upstream: map[string]tasks.Result{
					"Python/Env/Version": {
						Status:  tasks.Info,
						Summary: "SUMMARY",
						Payload: []string{"3.9"},
					},
					"Python/Agent/Version": {
						Status:  tasks.Info,
						Summary: "SUMMARY",
						Payload: "5.20.4",
					},
				},
			},
			want: tasks.Result{
				Status:  tasks.Success,
				Summary: "Your 3.9 Python version(s) are supported by the Python Agent.",
			},
		},
		{
			name: "Multiple Python Versions are compatible",
			tr:   PythonRequirementsPythonVersion{},
			args: args{
				options: tasks.Options{
					Options: map[string]string{"Option1": "option"},
				},
				upstream: map[string]tasks.Result{
					"Python/Env/Version": {
						Status:  tasks.Info,
						Summary: "SUMMARY",
						Payload: []string{"3.9", "3.8"},
					},
					"Python/Agent/Version": {
						Status:  tasks.Info,
						Summary: "SUMMARY",
						Payload: "5.20.4",
					},
				},
			},
			want: tasks.Result{
				Status:  tasks.Success,
				Summary: "Your 3.9,3.8 Python version(s) are supported by the Python Agent.",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := PythonRequirementsPythonVersion{}
			if got := tr.Execute(tt.args.options, tt.args.upstream); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PythonRequirementsPythonVersion.Execute() = %v, want %v", got, tt.want)
			}
		})
	}
}
