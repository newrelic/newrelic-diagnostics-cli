package env

import (
	"errors"
	"fmt"
	"runtime"
	"strings"
	"testing"

	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks/base/config"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/shirou/gopsutil/process"
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
		if runtime.GOOS != "windows" {
			Context("when it finds a java process that has the Java agent attached to it", func() {
				BeforeEach(func() {
					expectedPayload = []ProcIdAndArgs{}
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
					p.getCwd = func(process.Process) (string, error) {
						return "/root/go/src/github.com/newrelic/newrelic-diagnostics-cli", nil
					}
					expectedPayload = append(expectedPayload, ProcIdAndArgs{Proc: javaProcesses[0], CmdLineArgs: cmdLineArgsList, Cwd: "/root/go/src/github.com/newrelic/newrelic-diagnostics-cli", JarPath: "/root/go/src/github.com/newrelic/newrelic-diagnostics-cli/newrelic.jar", EnvVars: envVarsPayload})
				})

				It("should return a tasks.Result with a success status, a summary and a payload", func() {
					successSummary := fmt.Sprintf("We detected %d New Relic Java Agent(s) running on this host.", len(expectedPayload))
					Expect(result.Status).To(Equal(tasks.Success))
					Expect(result.Summary).To(Equal(successSummary))
					Expect(result.Payload).To(Equal(expectedPayload))
				})
			})

			Context("when it finds a java process that has the Java agent attached and the .jar file has been renamed to include a version in the name", func() {
				BeforeEach(func() {
					expectedPayload = []ProcIdAndArgs{}
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
					cmdLineArgs := "-javaagent:/root/go/src/github.com/newrelic/newrelic-diagnostics-cli/newrelic-1.8.0.jar"
					p.getCmdLineArgs = func(process.Process) (string, error) {
						return cmdLineArgs, nil
					}
					cmdLineArgsList := strings.Split(cmdLineArgs, " ")
					p.getCwd = func(process.Process) (string, error) {
						return "/root/go/src/github.com/newrelic/newrelic-diagnostics-cli", nil
					}
					expectedPayload = append(expectedPayload, ProcIdAndArgs{Proc: javaProcesses[0], CmdLineArgs: cmdLineArgsList, Cwd: "/root/go/src/github.com/newrelic/newrelic-diagnostics-cli", JarPath: "/root/go/src/github.com/newrelic/newrelic-diagnostics-cli/newrelic-1.8.0.jar", EnvVars: envVarsPayload})
				})

				It("should return a tasks.Result with a success status, a summary and a payload", func() {
					successSummary := fmt.Sprintf("We detected %d New Relic Java Agent(s) running on this host.", len(expectedPayload))
					Expect(result.Status).To(Equal(tasks.Success))
					Expect(result.Summary).To(Equal(successSummary))
					Expect(result.Payload).To(Equal(expectedPayload))
				})
			})

			Context("when it finds a java process but can't find the java agent", func() {
				BeforeEach(func() {
					expectedPayload = []ProcIdAndArgs{}
					options = tasks.Options{}
					envVarsPayload := map[string]string{
						"HOME": "/Users/shuayhuaca",
						"PATH": "/usr/local/opt/ruby/bin:/Users/shuayhuaca/.nvm/versions/node/v8.16.0/bin:/opt/apache-maven/bin/:/opt/apache-maven/bin:/usr/local/bin:/usr/bin:/bin:/usr/sbin:/sbin:/Applications/VMware Fusion.app/Contents/Public:/usr/local/go/bin:/usr/local/MacGPG2/bin:/Users/shuayhuaca/desktop/scripts:/usr/local/bin:/usr/bin:/bin:/usr/sbin:/sbin:/Applications/VMware Fusion.app/Contents/Public:/usr/local/go/bin:/usr/local/MacGPG2/bin:/Users/shuayhuaca/desktop/projects/nand2tetris/tools:/usr/local/go/bin:/Users/shuayhuaca/go/bin:/Applications/Visual Studio Code.app/Contents/Resources/app/bin:/Users/shuayhuaca/.rvm/bin",
					}
					upstream = map[string]tasks.Result{
						"Java/Config/Agent": {
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
						"Base/Env/CollectEnvVars": {
							Status:  tasks.Success,
							Payload: envVarsPayload,
						},
					}
					javaProcesses := []process.Process{
						{
							Pid: 1,
						},
					}
					p.findProcByName = func(string) ([]process.Process, error) {
						return javaProcesses, nil
					}
					cmdLineArgs := "-javaagent:/root/go/src/github.com/newrelic/newrelic-diagnostics-cli/bluerelic-1.8.0.jar"
					p.getCmdLineArgs = func(process.Process) (string, error) {
						return cmdLineArgs, nil
					}
				})

				It("should return a tasks.Result with a failure status and a summary of the failure", func() {
					Expect(result.Status).To(Equal(tasks.Failure))
					Expect(result.Summary).To(Equal("Failed to find the New Relic Java Agent Jar in the following JVM argument: " + "-javaagent:/root/go/src/github.com/newrelic/newrelic-diagnostics-cli/bluerelic-1.8.0.jar" + "\nIf this is another Java Agent, keep in mind that New Relic is not compatible with other additional agents."))
				})
			})
		}
	})

	Describe("getJarInfoFromCmdLineArgs", func() {
		command := "-javaagent:"
		isWindows := runtime.GOOS == "windows"

		if !isWindows {
			Context("When given a cmdLineArgsStr that contains a path and an agent name of newrelic.jar", func() {
				It("should return the path and no error", func() {
					path := "/Users/shuayhuaca/Desktop/projects/java/luces/"
					javaAgent := "newrelic.jar"
					commandLineArgsStr := command + path + javaAgent
	
					resultPath, resultFilename, err := getJarInfoFromCmdLineArgs(commandLineArgsStr)
					Expect(resultPath).To(Equal(path))
					Expect(resultFilename).To(Equal(javaAgent))
					Expect(err).To(BeNil())
				})
			})
			
			Context("When given a cmdLineArgsStr that contains a path and an agent name of newrelic-1.8.3.jar", func() {
				It("should return the path and no error", func() {
					path := "/Users/shuayhuaca/Desktop/projects/java/luces/"
					javaAgent := "newrelic-1.8.3.jar"
					commandLineArgsStr := command + path + javaAgent
	
					resultPath, resultFilename, err := getJarInfoFromCmdLineArgs(commandLineArgsStr)
					Expect(resultPath).To(Equal(path))
					Expect(resultFilename).To(Equal(javaAgent))
					Expect(err).To(BeNil())
				})
			})
	
			Context("When cmdLineArgsStr content contains things other than the javaagent argument", func() {
				It("should return the path and no error", func() {
					otherBefore := "/usr/bin/java -Dnewrelic.logfile=/Users/shuayhuaca/Desktop/newrelic_agent.log -jar "
					path := "/Users/shuayhuaca/Desktop/projects/java/luces/"
					javaAgent := "newrelic-1.8.3.jar"
					otherAfter := " build/libs/lucessqs-1.0-SNAPSHOT.jar"
					commandLineArgsStr := otherBefore + command + path + javaAgent + otherAfter
	
					resultPath, resultFilename, err := getJarInfoFromCmdLineArgs(commandLineArgsStr)
					Expect(resultPath).To(Equal(path))
					Expect(resultFilename).To(Equal(javaAgent))
					Expect(err).To(BeNil())
				})
			})
	
			Context("When given a cmdLineArgsStr that does not contain a path, but has a valid agent name", func() {
				It("should return ./ as the path and no error", func() {
					javaAgent := "newrelic-1.8.3.jar"
					commandLineArgsStr := command + javaAgent
	
					resultPath, resultFilename, err := getJarInfoFromCmdLineArgs(commandLineArgsStr)
					Expect(resultPath).To(Equal("./"))
					Expect(resultFilename).To(Equal(javaAgent))
					Expect(err).To(BeNil())
				})
			})
	
			Context("When given a cmdLineArgsStr with an invalid agent name", func() {
				It("should return the path and an error", func() {
					path := "/Users/shuayhuaca/Desktop/projects/java/luces/"
					javaAgent := "bluerelic.jar"
					commandLineArgsStr := command + path + javaAgent
	
					resultPath, resultFilename, err := getJarInfoFromCmdLineArgs(commandLineArgsStr)
					Expect(resultPath).To(Equal(path))
					Expect(resultFilename).To(Equal(javaAgent))
					Expect(err.Error()).To(Equal(commandLineArgsStr))
				})
			})
		}

		if isWindows {
			Context("When given a cmdLineArgsStr using a windows path", func() {
				It("should return the path and no error", func() {
					path := `\newrelic\`
					javaAgent := "newrelic.jar"
					commandLineArgsStr := command + path + javaAgent
	
					resultPath, resultFilename, err := getJarInfoFromCmdLineArgs(commandLineArgsStr)
					Expect(resultPath).To(Equal(path))
					Expect(resultFilename).To(Equal(javaAgent))
					Expect(err).To(BeNil())
				})
			})

			Context("When given a cmdLineArgsStr using a windows path where jar does not have new relic in the name", func() {
				It("should return the path and an error", func() {
					path := `C:\projects\myapp\newrelic\`
					javaAgent := "bluerelic.jar"
					commandLineArgsStr := command + path + javaAgent
	
					resultPath, resultFilename, err := getJarInfoFromCmdLineArgs(commandLineArgsStr)
					Expect(resultPath).To(Equal(path))
					Expect(resultFilename).To(Equal(javaAgent))
					Expect(err.Error()).To(Equal(commandLineArgsStr))
				})
			})
		}
		
	})
})
