package agent

import (
	"testing"

	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestIOSAgentVersion(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "iOS/Agent/Version test suite")
}

var _ = Describe("iOS/Agent/Version", func() {
	var p iOSAgentVersion

	Describe("Identifier()", func() {
		It("Should return correct identifier", func() {
			expectedIdentifier := tasks.Identifier{
				Category:    "iOS",
				Subcategory: "Agent",
				Name:        "Version",
			}

			Expect(p.Identifier()).To(Equal(expectedIdentifier))
		})

	})

	Describe("Explain()", func() {
		It("Should return correct task explanation string", func() {
			expectedExaplanation := "Determine New Relic iOS agent version"
			Expect(p.Explain()).To(Equal(expectedExaplanation))
		})
	})

	Describe("Dependencies()", func() {
		It("Should return an expected slice of dependencies", func() {
			expectedDependencies := []string{
				"Base/Config/Collect",
				"iOS/Env/Detect",
			}
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
					"iOS/Env/Detect": tasks.Result{
						Status: tasks.Failure,
					},
				}
			})

			It("should return an expected none result status", func() {
				Expect(result.Status).To(Equal(tasks.None))
			})

			It("should return an expected none result summary", func() {
				Expect(result.Summary).To(Equal("Task did not meet requirements necessary to run: iOS environment not detected."))
			})
		})

		Context("upstream dependency returned unexpected payload type", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Base/Config/Collect": tasks.Result{
						Status:  tasks.Success,
						Payload: "This should be of type []config.ConfigElement, but it's a string instead",
					},
					"iOS/Env/Detect": tasks.Result{
						Status: tasks.Info,
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
})
