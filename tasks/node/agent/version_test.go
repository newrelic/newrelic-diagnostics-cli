package agent

import (
	"testing"

	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
	NodeEnv "github.com/newrelic/newrelic-diagnostics-cli/tasks/node/env"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestNodeAgentVersion(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Node/Agent test suite")
}

var _ = Describe("Node/Agent/Version", func() {
	var p NodeAgentVersion
	Describe("getNodeVerFromPayload()", func() {
		var (
			incomingPayload []NodeEnv.NodeModuleVersion
			output          string
		)

		JustBeforeEach(func() {
			output = getNodeVerFromPayload(incomingPayload)
		})

		Context("When given a payload containing a Node Agent version", func() {
			BeforeEach(func() {
				incomingPayload = []NodeEnv.NodeModuleVersion{
					{
						Module:  "Apollo",
						Version: "1.2.3",
					},
					{
						Module:  "newrelic",
						Version: "2.1.4",
					},
				}
			})
			It("Should return the expected agent version", func() {
				expectedReturn := "newrelic 2.1.4"
				Expect(output).To(Equal(expectedReturn))
			})
		})

		Context("When given a payload which does not contain a Node Agent Version", func() {
			BeforeEach(func() {
				incomingPayload = []NodeEnv.NodeModuleVersion{
					{
						Module:  "Apollo",
						Version: "1.2.3",
					},
					{
						Module:  "lodash",
						Version: "4.2.03",
					},
				}
			})
			It("Should return an empty Node agent version", func() {
				expectedReturn := ""
				Expect(output).To(Equal(expectedReturn))
			})
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
		Context("When given a Node Agent Module isn't found", func() {
			BeforeEach(func() {
				upstream = map[string]tasks.Result{
					"Node/Env/Dependencies": tasks.Result{
						Status: tasks.Info,
						Payload: []NodeEnv.NodeModuleVersion{
							{
								Module:  "Apollo",
								Version: "1.2.3",
							},
							{
								Module:  "lodash",
								Version: "4.2.03",
							},
						},
					},
					"Node/Config/Agent": tasks.Result{
						Status: tasks.Success,
					},
				}
			})
			It("Should return a tasks.Warning", func() {
				expectedReturn := tasks.Result{
					Status:  tasks.Warning,
					Summary: "Node Agent Module not found for newrelic",
				}
				Expect(result).To(Equal(expectedReturn))
			})
		})

	})
})
