// +build linux darwin

package env

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/shirou/gopsutil/process"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks/base/config"
)

func TestJavaEnvProcess(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Java/Env/* test suite")
}

var _ = Describe("JavaEnvProcess", func() {
	var p JavaEnvProcess

	Describe("Execute()", func() {
		var (
			options  tasks.Options
			upstream map[string]tasks.Result
			result   tasks.Result
		)
		expectedPayload := []ProcIdAndArgs{}
		JustBeforeEach(func() {
			result = p.Execute(options, upstream)
		})
		Context("When there is no Java agent config file found", func() {
			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Java/Config/Agent": tasks.Result{
						Status: tasks.None,
					},
				}
			})
			It("should return a task Result with Status None and a Summary", func() {
				Expect(result.Status).To(Equal(tasks.None))
				Expect(result.Summary).To(Equal("Java agent config file was not detected on this host. This task did not run"))
			})

		})
		Context("when we encounter an error when looking for Java processes", func() {
			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Java/Config/Agent": tasks.Result{
						Status:  tasks.Success,
						Summary: "Java agent identified as present on system",
						Payload: []config.ValidateElement{
							{
								Config: config.ConfigElement{
									FileName: "newrelic.yml",
									FilePath: "fixtures/java/newrelic/",
								},
							},
						},
					},
					"Base/Env/CollectEnvVars": tasks.Result{
						Status:  tasks.Success,
						Payload: map[string]string{},
					},
				}
				p.findProcByName = func(string) ([]process.Process, error) {
					return []process.Process{}, errors.New("an error message")
				}
			})

			It("should return an tasks result with a error status and a summary", func() {
				Expect(result.Status).To(Equal(tasks.Error))
				Expect(result.Summary).To(Equal("We encountered an error while detecting all running Java processes: an error message"))
			})
		})
		Context("when it finds a java process that has the Java agent attached to it", func() {
			BeforeEach(func() {
				options = tasks.Options{}
				envVarsPayload := map[string]string{
					"HOME": "/Users/shuayhuaca",
					"PATH": "/usr/local/opt/ruby/bin:/Users/shuayhuaca/.nvm/versions/node/v8.16.0/bin:/opt/apache-maven/bin/:/opt/apache-maven/bin:/usr/local/bin:/usr/bin:/bin:/usr/sbin:/sbin:/Applications/VMware Fusion.app/Contents/Public:/usr/local/go/bin:/usr/local/MacGPG2/bin:/Users/shuayhuaca/desktop/scripts:/usr/local/bin:/usr/bin:/bin:/usr/sbin:/sbin:/Applications/VMware Fusion.app/Contents/Public:/usr/local/go/bin:/usr/local/MacGPG2/bin:/Users/shuayhuaca/desktop/projects/nand2tetris/tools:/usr/local/go/bin:/Users/shuayhuaca/go/bin:/Applications/Visual Studio Code.app/Contents/Resources/app/bin:/Users/shuayhuaca/.rvm/bin",
				}
				upstream = map[string]tasks.Result{
					"Java/Config/Agent": tasks.Result{
						Status:  tasks.Success,
						Summary: "Java agent identified as present on system",
						Payload: []config.ValidateElement{
							{
								Config: config.ConfigElement{
									FileName: "newrelic.yml",
									FilePath: "fixtures/java/newrelic/",
								},
							},
						},
					},
					"Base/Env/CollectEnvVars": tasks.Result{
						Status:  tasks.Success,
						Payload: envVarsPayload,
					},
				}
				javaProcesses := []process.Process{
					process.Process{
						Pid: 1,
					},
				}
				p.findProcByName = func(string) ([]process.Process, error) {
					return javaProcesses, nil
				}
				cmdLineArgs := "-javaagent:/root/go/src/github.com/newrelic/newrelic-diagnostics-cli/newrelic.jar"
				p.getCmdLineArgs = func(process.Process) (string, error) {
					return cmdLineArgs, nil
				}
				cmdLineArgsList := strings.Split(cmdLineArgs, " ")
				p.getCurrentDir = func(process.Process, string) string {
					return "/root/go/src/github.com/newrelic/newrelic-diagnostics-cli"
				}
				expectedPayload = append(expectedPayload, ProcIdAndArgs{Proc: javaProcesses[0], CmdLineArgs: cmdLineArgsList, Cwd: "/root/go/src/github.com/newrelic/newrelic-diagnostics-cli", EnvVars: envVarsPayload})
			})

			It("should return a tasks.Result with a success status, a summary and a payload", func() {
				successSummary := fmt.Sprintf("We detected %d New Relic Java Agent(s) running on this host.", len(expectedPayload))
				Expect(result.Status).To(Equal(tasks.Success))
				Expect(result.Summary).To(Equal(successSummary))
				Expect(result.Payload).To(Equal(expectedPayload))
			})
		})
	})
})
