package agent

import (
	"testing"

	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
	BrowserFixtures "github.com/newrelic/newrelic-diagnostics-cli/tasks/fixtures/Browser"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestBrowserAgentDetect(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Browser/Agent/* test suite")
}

var _ = Describe("Browser/Agent/Detect", func() {
	var p BrowserAgentDetect

	Describe("Identifier()", func() {
		It("Should return correct identifier", func() {
			expectedIdentifier := tasks.Identifier{
				Category:    "Browser",
				Subcategory: "Agent",
				Name:        "Detect",
			}

			Expect(p.Identifier()).To(Equal(expectedIdentifier))
		})
	})

	Describe("Explain()", func() {
		It("Should return correct string", func() {
			expectedString := "Detect New Relic Browser agent from provided URL"

			Expect(p.Explain()).To(Equal(expectedString))
		})
	})

	Describe("Dependencies()", func() {
		It("Should return correct slice", func() {
			expectedDependencies := []string{"Browser/Agent/GetSource"}

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

		Context("when source loader of browser agent is detected in head", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Browser/Agent/GetSource": tasks.Result{
						Status: tasks.Success,
						Payload: BrowserAgentSourcePayload{
							URL:    "https://www.testingmctestface.com",
							Source: BrowserFixtures.HTMLWithGoodLoader,
							Loader: []string{BrowserFixtures.AgentScript},
						},
					},
				}
			})

			It("should return a success status", func() {
				Expect(result.Status).To(Equal(tasks.Success))
			})

			It("should return an expected none result summary", func() {
				Expect(result.Summary).To(Equal("Found values for browser agent"))
			})
		})
	})

})
