package agent

/*
import (
	"testing"

	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	BrowserResponse "github.com/newrelic/newrelic-diagnostics-cli/tasks/fixtures/Browser"
	AgentScript "github.com/newrelic/newrelic-diagnostics-cli/tasks/fixtures/Browser"
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
						Status:  tasks.Success,
						Payload: BrowserAgentSourcePayload {
							URL: "https://www.testingmctestface.com",
							// TODO: Check with team to verify using real url in test mock is good practice
							Source: BrowserResponse,
							Loader: []string{AgentScript},
						},
					},
				},
			})

			It("should return a success status", func() {
				Expect(result.Status).To(Equal(tasks.Success))
			})

			It("should return an expected none result summary", func() {
				Expect(result.Summary).To(Equal("Found values for browser agent"))
			})
		})
	})


			Context("when upstream dependency does not return successful status", func() {

				BeforeEach(func() {
					options = tasks.Options{}
					upstream = map[string]tasks.Result{
						"Node/Env/Version": tasks.Result{
							Status:  tasks.Error,
							Payload: "4",
						},
					}
					p = NodeEnvVersionCompatibility{
						supportedNodeVersions: []string{"6.0+"},
					}
				})

				It("should return an expected none result status", func() {
					Expect(result.Status).To(Equal(tasks.None))
				})

				It("should return an expected none result summary", func() {
					Expect(result.Summary).To(Equal("Task did not meet requirements necessary to run: Node is not installed"))
				})
			})

			Context("when running Ver.IsCompatible returns an error", func() {

				BeforeEach(func() {
					options = tasks.Options{}
					parsedVer, _ := tasks.ParseVersion("6.0.1.2")
					upstream = map[string]tasks.Result{
						"Node/Env/Version": tasks.Result{
							Status:  tasks.Info,
							Payload: parsedVer,
						},
					}
					p = NodeEnvVersionCompatibility{
						supportedNodeVersions: []string{"I Am Not AVersionRequirement"},
					}

				})

				It("should return an expected error result status", func() {
					Expect(result.Status).To(Equal(tasks.Error))
				})

				It("should return an expected error result summary", func() {
					Expect(result.Summary).To(Equal("There was an issue when checking for Node.js Version compatibility: Unable to parse version: I Am Not AVersionRequirement"))
				})
			})

			Context("When running versionIsCompatible returns false", func() {

				BeforeEach(func() {
					options = tasks.Options{}
					parsedVer, _ := tasks.ParseVersion("5.0.7")
					upstream = map[string]tasks.Result{
						"Node/Env/Version": tasks.Result{
							Status:  tasks.Info,
							Payload: parsedVer,
						},
					}
					p = NodeEnvVersionCompatibility{
						supportedNodeVersions: []string{"6.0+"},
					}
				})

				It("should return an expected failure result status", func() {
					Expect(result.Status).To(Equal(tasks.Failure))
				})
				It("should return an expected failure result summary", func() {
					Expect(result.Summary).To(Equal("Your current Node.js version, 5.0.7.0, is not compatible with New Relic's Node.js agent"))
				})
			})
			Context("When running versionIsCompatible returns true", func() {

				BeforeEach(func() {
					options = tasks.Options{}
					parsedVer, _ := tasks.ParseVersion("10.0.7")
					upstream = map[string]tasks.Result{
						"Node/Env/Version": tasks.Result{
							Status:  tasks.Info,
							Payload: parsedVer,
						},
					}
				})
				p = NodeEnvVersionCompatibility{
					supportedNodeVersions: []string{"6.0+"},
				}

				It("should return an expected success result status", func() {
					Expect(result.Status).To(Equal(tasks.Success))
				})
				It("should return an expected succes result summary", func() {
					Expect(result.Summary).To(Equal("Your current Node.js version, 10.0.7.0, is compatible with New Relic's Node.js agent"))
				})
			})
		})

})
*/
