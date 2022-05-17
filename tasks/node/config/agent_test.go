package config

import (
	"testing"

	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks/base/config"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestNodeEnvOsCheck(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Node/Agent/* test suite")
}

var _ = Describe("Node/Config/Agent", func() {

	var p NodeConfigAgent

	Describe("Dependencies()", func() {
		It("Should return an slice required to run this task", func() {
			expectedDependencies := []string{"Base/Env/CollectEnvVars",
				"Base/Config/Collect",
				"Base/Config/Validate"}
			Expect(p.Dependencies()).To(Equal(expectedDependencies))
		})
	})
	Describe("Execute()", func() {
		var (
			//inputs
			options  tasks.Options
			upstream map[string]tasks.Result
			//output
			result tasks.Result
		)
		JustBeforeEach(func() {
			result = p.Execute(options, upstream)
		})
		Context("When node agent validation is available", func() {
			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Base/Env/CollectEnvVars": tasks.Result{
						Status: tasks.Info,
						Payload: map[string]string{
							"NEW_RELIC_APP_NAME": "luces-app",
						},
					},
					"Base/Config/Validate": tasks.Result{
						Status: tasks.Success,
						Payload: []config.ValidateElement{
							{
								Config: config.ConfigElement{
									FileName: "newrelic.js",
									FilePath: "/newrelic/newrelic.js",
								},
								ParsedResult: tasks.ValidateBlob{
									Key:      "",
									Path:     "",
									RawValue: nil,
									Children: []tasks.ValidateBlob{
										tasks.ValidateBlob{
											Key:      "app_name",
											Path:     "",
											RawValue: "luces-app",
											Children: nil,
										},
									},
								},
							},
						},
					},
				}
			})
			It("Should return a None result status", func() {
				Expect(result.Status).To(Equal(tasks.Success))
			})
			It("Should return a None result summary", func() {
				Expect(result.Summary).To(Equal("Node agent identified as present on system"))
			})
		})
	})
})
