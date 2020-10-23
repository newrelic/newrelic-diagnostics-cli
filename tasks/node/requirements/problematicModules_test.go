package requirements

import (
	"fmt"
	"testing"

	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
	dependencies "github.com/newrelic/newrelic-diagnostics-cli/tasks/node/env"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestNodeRequirementsProblematicModules(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Node/Requirements/* test suite")
}

var _ = Describe("Node/Requirements/ProblematicModules", func() {
	var p NodeRequirementsProblematicModules
	Describe("Identifier()", func() {
		It("should return expected identifier", func() {
			expectedIdentifier := tasks.Identifier{
				Name:        "ProblematicModules",
				Category:    "Node",
				Subcategory: "Requirements",
			}
			Expect(p.Identifier()).To(Equal(expectedIdentifier))
		})
	})
	Describe("Explain()", func() {
		It("should return the correct explanation", func() {
			Expect(p.Explain()).To(Equal("This task declares unsupported Node Agent technologies"))
		})
	})
	Describe("Dependencies()", func() {
		It("should return expected dependencies slice", func() {
			Expect(p.Dependencies()).To(Equal(
				[]string{
					"Node/Env/Dependencies",
				},
			))
		})
	})
	Describe("Execute()", func() {
		var (
			options  tasks.Options
			upstream map[string]tasks.Result
			result   tasks.Result
		)
		JustBeforeEach(func() {
			result = p.Execute(options, upstream)
		})
		Context("when Node/Env/Dependencies did not return an Info status", func() {
			BeforeEach(func() {
				upstream = map[string]tasks.Result{
					"Node/Env/Dependencies": tasks.Result{
						Status: tasks.None,
					},
				}
			})
			It("should return an expected None Status", func() {
				Expect(result.Status).To(Equal(tasks.None))
			})
			It("should return an expected summary", func() {
				Expect(result.Summary).To(Equal("A list of Node modules was not found. This task did not run"))
			})
		})
		Context("when using React, Babel and Next framework", func() {
			BeforeEach(func() {
				upstream = map[string]tasks.Result{
					"Node/Env/Dependencies": tasks.Result{
						Status: tasks.Info,
						Payload: []dependencies.NodeModuleVersion{
							{
								Module:  "react",
								Version: "1.0.0",
							},
							{
								Module:  "react-dom",
								Version: "1.0.0",
							},
							{
								Module:  "@types/react-dom",
								Version: "16.9.4",
							},
							{
								Module:  "next",
								Version: "8.7.0",
							},
							{
								Module:  "@newrelic/native-metrics",
								Version: "5.1.0",
							},
							{
								Module:  "@babel/generator",
								Version: "7.8.3",
							},
						},
					},
				}
			})
			It("should return an expected None Status", func() {
				Expect(result.Status).To(Equal(tasks.Warning))
			})
			It("should return an expected summary", func() {
				Expect(result.Summary).To(Equal("- You are not using a supported framework by the Node Agent. In order to get monitoring data, you'll have to apply manual instrumentation using our APIs. For more information: https://docs.newrelic.com/docs/agents/nodejs-agent/supported-features/nodejs-custom-instrumentation\n- We noticed that you are using: react, react-dom. If you are looking to monitor a client side app, beware that the Node Agent only monitors server side frameworks. To get metrics for front-end libraries/frameworks use the Browser Agent instead: https://docs.newrelic.com/docs/browser/new-relic-browser/getting-started/compatibility-requirements-new-relic-browser\n- We have detected the following unsupported module(s) in your application: @babel/generator. This may cause instrumentation issues and inconsistency of data for the Node Agent.\n"))
			})
		})
		Context("when we do not find a supported framework among their node modules and their are using problematic modules as well", func() {
			BeforeEach(func() {
				upstream = map[string]tasks.Result{
					"Node/Env/Dependencies": tasks.Result{
						Status: tasks.Info,
						Payload: []dependencies.NodeModuleVersion{
							{
								Module:  "sails",
								Version: "1.2.4",
							},
							{
								Module:  "@babel/node",
								Version: "7.2.2",
							},
							{
								Module:  "@babel/core",
								Version: "7.2.2",
							},
						},
					},
				}
			})
			It("should return an expected None Status", func() {
				Expect(result.Status).To(Equal(tasks.Warning))
			})
			It("should return an expected summary", func() {
				fmt.Println("MY SUMARY: ", result.Summary)
				Expect(result.Summary).To(Equal("- You are not using a supported framework by the Node Agent. In order to get monitoring data, you'll have to apply manual instrumentation using our APIs. For more information: https://docs.newrelic.com/docs/agents/nodejs-agent/supported-features/nodejs-custom-instrumentation\n- We have detected the following unsupported module(s) in your application: @babel/node, @babel/core. This may cause instrumentation issues and inconsistency of data for the Node Agent.\n- Keep in mind that if you are looking for additional Node.js runtime level statistics, you'll need to install our optional module: @newrelic/native-metrics. For more information: https://docs.newrelic.com/docs/agents/nodejs-agent/supported-features/nodejs-vm-measurements\n"))
			})
		})

	})
})
