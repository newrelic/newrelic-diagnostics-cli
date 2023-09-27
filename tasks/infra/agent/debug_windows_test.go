package agent

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

var _ = Describe("Infra/Agent/Debug", func() {

	var p InfraAgentDebug

	Describe("Execute()", func() {
		var (
			result   tasks.Result
			options  tasks.Options
			upstream map[string]tasks.Result
		)

		JustBeforeEach(func() {
			result = p.Execute(options, upstream)
		})

		Context("When Running on Windows", func() {
			Describe("and the infra version is not supported", func() {
				BeforeEach(func() {
					options = tasks.Options{}
					upstream = map[string]tasks.Result{
						"Infra/Log/Collect": {
							Status:  tasks.Success,
							Payload: []string{"test"},
						},
						"Infra/Agent/Version": {
							Status: tasks.Info,
							Payload: tasks.Ver{
								Major: 1,
								Minor: 5,
								Patch: 0,
								Build: 0,
							},
						},
						"Base/Env/CollectEnvVars": {
							Status: tasks.None,
						},
					}
				})
				It("should return an expected result status", func() {
					Expect(result.Status).To(Equal(tasks.Failure))
				})

				It("should return an expected result summary", func() {
					Expect(result.Summary).To(Equal("Infrastructure debug CTL binary not available in detected version of Infrastructure Agent(1.5.0.0). Minimum required Infrastructure Agent version is: 1.7.0.0"))
				})
			})
			Describe("and the infra version is supported", func() {
				BeforeEach(func() {

					options = tasks.Options{}
					upstream = map[string]tasks.Result{
						"Infra/Log/Collect": {
							Status:  tasks.Success,
							Payload: []string{"test"},
						},
						"Infra/Agent/Version": {
							Status: tasks.Info,
							Payload: tasks.Ver{
								Major: 1,
								Minor: 7,
								Patch: 1,
								Build: 0,
							},
						},
						"Base/Env/CollectEnvVars": {
							Status: tasks.Info,
							Payload: map[string]string{
								"ProgramFiles": `C:\Program Files`,
							},
						},
					}
					p.blockWithProgressbar = func(int) {}
					p.cmdExecutor = func(string, ...string) ([]byte, error) { return []byte{}, nil }
				})
				It("should return an expected result status", func() {
					Expect(result.Status).To(Equal(tasks.Success))
				})

				It("should return an expected result summary", func() {
					Expect(result.Summary).To(Equal("Successfully enabled New Relic Infrastructure debug logging with newrelic-infra-ctl"))
				})
			})
		})

	})
})
