package requirements

import (
	"testing"

	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Python/Requirements/Webframework", func() {
	var p PythonRequirementsWebframework

	Describe("Identifier()", func() {
		It("Should return correct identifier", func() {
			expectedIdentifier := tasks.Identifier{
				Category:    "Python",
				Subcategory: "Requirements",
				Name:        "Webframework",
			}

			Expect(p.Identifier()).To(Equal(expectedIdentifier))
		})

	})

	Describe("Explain()", func() {
		It("Should return correct task explanations string", func() {
			expectedExplanation := "Check web framework compatibility with New Relic Python agent"
			Expect(p.Explain()).To(Equal(expectedExplanation))
		})
	})

	Describe("Dependencies()", func() {
		It("Should return an expected slice of dependencies", func() {
			expectedDependencies := []string{"Python/Env/Dependencies"}
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
					"Python/Env/Dependencies": {
						Status: tasks.Failure,
					},
				}
			})

			It("should return an expected none result status", func() {
				Expect(result.Status).To(Equal(tasks.None))
			})

			It("should return an expected none result summary", func() {
				Expect(result.Summary).To(Equal("No dependencies found. This task didn't run."))
			})
		})

		Context("upstream dependency task returned unexpected Type", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Python/Env/Dependencies": {
						Status:  tasks.Info,
						Payload: []int{1, 2, 3},
					},
				}
			})

			It("should return an expected none result status", func() {
				Expect(result.Status).To(Equal(tasks.Error))
			})

			It("should return an expected none result summary", func() {
				Expect(result.Summary).To(Equal(tasks.AssertionErrorSummary))
			})
		})

	})

	Describe("extractFrameworkDetails", func() {
		It("It should strip the alphabetic characters from version", func() {
			pipFreezeOutputItem := "flask==0.1a2dev"
			framework, version := extractFrameworkDetails(pipFreezeOutputItem)
			Expect(version).To(Equal("0.12"))
			Expect(framework).To(Equal("flask"))
		})

	})

})

func Test_checkWebframework(t *testing.T) {
	webframeworkTests := []struct {
		pipFreezeOutputItem []string
		want                tasks.Status
	}{

		{pipFreezeOutputItem: []string{"flask==0.12"}, want: tasks.Success},
		{pipFreezeOutputItem: []string{"sanic==2.0"}, want: tasks.Success},
		{pipFreezeOutputItem: []string{"falcon==2.0"}, want: tasks.Success},
		{pipFreezeOutputItem: []string{"aiohttp==3.1.3"}, want: tasks.Success},
		{pipFreezeOutputItem: []string{"tornado==6"}, want: tasks.Success},
		{pipFreezeOutputItem: []string{"aiohttp==2.0.7", "flask==0.12"}, want: tasks.Success},
		{pipFreezeOutputItem: []string{"twisted==18.4.0"}, want: tasks.Warning},
		{pipFreezeOutputItem: []string{"aiohttp==2.0.7"}, want: tasks.Warning},
		{pipFreezeOutputItem: []string{"tornado==5.0"}, want: tasks.Warning},
		{pipFreezeOutputItem: []string{"aiohttp==2.0.7", "twisted==18.4.0"}, want: tasks.Warning},
	}
	for _, webframeworkTest := range webframeworkTests {
		webframework := checkWebframework(webframeworkTest.pipFreezeOutputItem)
		if webframeworkTest.want != webframework.Status {
			t.Errorf("checkWebframework() = Task Result %v, framework tested %v, want %v", webframework, webframeworkTest.pipFreezeOutputItem, webframeworkTest.want)
		}
	}
}
