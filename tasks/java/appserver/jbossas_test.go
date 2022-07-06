package appserver

import (
	"errors"
	"os"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/shirou/gopsutil/v3/process"

	baseConfig "github.com/newrelic/newrelic-diagnostics-cli/config"
	tasks "github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

func TestMain(m *testing.M) {
	//Toggle to enable verbose logging
	baseConfig.LogLevel = baseConfig.Info
	os.Exit(m.Run())
}

func TestCollector(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "JBoss test Suite")
}

var _ = Describe("JavaAppserverJBossAsCheck", func() {

	var (
		p        JavaAppserverJBossAsCheck
		upstream map[string]tasks.Result
		result   tasks.Result
	)

	Describe("Identifier", func() {
		It("Should return the identifier", func() {
			Expect(p.Identifier()).To(Equal(tasks.Identifier{Name: "JBossAsCheck", Category: "Java", Subcategory: "Appserver"}))
		})
	})
	Describe("Explain", func() {
		It("Should return explain", func() {
			Expect(p.Explain()).To(Equal("Check JBoss AS version compatibility with New Relic Java agent"))
		})
	})
	Describe("Dependencies", func() {
		It("Should return dependencies list", func() {
			Expect(p.Dependencies()).To(Equal([]string{"Base/Env/CollectEnvVars", "Java/Env/Process"}))
		})
	})
	Describe("Execute", func() {

		JustBeforeEach(func() {
			result = p.Execute(tasks.Options{}, upstream)
		})

		Context("When running on linux with JBOSS_HOME not set and no running Java processes", func() {
			BeforeEach(func() {
				p.findProcessByName = func(string) ([]process.Process, error) {
					return []process.Process{}, nil
				}

				p.findFiles = func([]string, []string) []string {
					return []string{"foo"}
				}

				p.returnSubstringInFile = func(string, string) ([]string, error) {
					return []string{}, errors.New("blargs")
				}
				upstream = map[string]tasks.Result{
					"Base/Env/CollectEnvVars": tasks.Result{},
					"Java/Env/Process": tasks.Result{
						Status: tasks.Success,
					},
				}
			})
			It("Should return fallthrough result", func() {
				Expect(result.Status).To(Equal(tasks.None))
				Expect(result.Summary).To(Equal("Could not find JBoss AS Home Path. Assuming JBoss AS is not installed"))
			})

		})

		Context("When running on linux and JBOSS_HOME ENV var not set and java process found", func() {
			BeforeEach(func() {
				p.getCmdline = func(process.Process) string {
					return "jboss.home.dir=/jboss-5.4.0/server"
				}
				p.findProcessByName = func(string) ([]process.Process, error) {
					return []process.Process{process.Process{Pid: 1}}, nil
				}
				upstream = map[string]tasks.Result{
					"Base/Env/CollectEnvVars": tasks.Result{},
					"Java/Env/Process": tasks.Result{
						Status: tasks.Success,
					},
				}
			})

			It("Should return Result a Successful result", func() {
				Expect(result.Status).To(Equal(tasks.Success))
			})

			It("Should indicate success in its Summary", func() {
				Expect(result.Summary).To(ContainSubstring("JBoss version supported"))
			})
		})

		Context("When running on windows and JBOSS_HOME ENV var not set and java process found", func() {
			BeforeEach(func() {
				p.getCmdline = func(process.Process) string {
					return `jboss.home.dir=C:\appserver\jboss`
				}
				p.findProcessByName = func(string) ([]process.Process, error) {
					return []process.Process{process.Process{Pid: 1}}, nil
				}
				p.findFiles = func([]string, []string) []string {
					return []string{"README.txt"}
				}
				p.returnSubstringInFile = func(string, string) ([]string, error) {
					return []string{"JBoss Application Server 5.4.0"}, nil
				}
			})
			It("Should return a Successful Result", func() {
				Expect(result.Status).To(Equal(tasks.Success))
			})
			It("Should indicate success in its Summary", func() {
				Expect(result.Summary).To(ContainSubstring("JBoss version supported"))
			})
		})

		Context("When error reading processes", func() {
			BeforeEach(func() {
				p.getCmdline = func(process.Process) string {
					return ""
				}
				p.findProcessByName = func(string) ([]process.Process, error) {
					return []process.Process{}, errors.New("could not read processes")
				}
			})

			It("Should return an error result", func() {
				Expect(result.Summary).To(Equal("Diagnostics CLI was unable to validate if your JBoss AS version is compatible with New Relic Java agent because it ran into an error when reading from your java process: could not read processes\nYou can take look at this documentation to verify if your version of JBoss is compatible: https://docs.newrelic.com/docs/agents/java-agent/getting-started/compatibility-requirements-java-agent#app-web-servers"))
				Expect(result.Status).To(Equal(tasks.Error))
			})
		})
		Context("When error retrieving list of processes", func() {
			BeforeEach(func() {
				p.findProcessByName = func(string) ([]process.Process, error) {
					return []process.Process{}, errors.New("i like sandwiches")
				}
				upstream = map[string]tasks.Result{
					"Base/Env/CollectEnvVars": tasks.Result{},
					"Java/Env/Process": tasks.Result{
						Status: tasks.Success,
					},
				}
			})
			It("Should return result from getAndParseJBossAsReadMeChecker ", func() {
				Expect(result.Summary).To(Equal("Diagnostics CLI was unable to validate if your JBoss AS version is compatible with New Relic Java agent because it ran into an error when reading from your java process: i like sandwiches\nYou can take look at this documentation to verify if your version of JBoss is compatible: https://docs.newrelic.com/docs/agents/java-agent/getting-started/compatibility-requirements-java-agent#app-web-servers"))
				Expect(result.Status).To(Equal(tasks.Error))
			})
		})
		Context("When jboss not detected as installed", func() {
			BeforeEach(func() {
				p.findProcessByName = func(string) ([]process.Process, error) {
					return []process.Process{}, nil
				}
				upstream = map[string]tasks.Result{
					"Base/Env/CollectEnvVars": tasks.Result{},
					"Java/Env/Process": tasks.Result{
						Status: tasks.Success,
					},
				}
			})
			It("Should return none result", func() {
				Expect(result.Summary).To(Equal("Could not find JBoss AS Home Path. Assuming JBoss AS is not installed"))
				Expect(result.Status).To(Equal(tasks.None))
			})
		})
		Context("When JBOSS_HOME is set and no readme files are found", func() {
			BeforeEach(func() {
				upstream = map[string]tasks.Result{
					"Base/Env/CollectEnvVars": {
						Status:  tasks.Info,
						Payload: map[string]string{"JBOSS_HOME": "/foo/bar"}},
					"Java/Env/Process": {
						Status: tasks.Success,
					},
				}

				p.findFiles = func([]string, []string) []string {
					return []string{}
				}
			})
			It("Should return error result", func() {
				Expect(result.Summary).To(Equal("Diagnostics CLI was unable to validate if your JBoss AS version is compatible with New Relic Java agent because it ran into an error when reading jboss readme: error finding JBoss version\nYou can take look at this documentation to verify if your version of JBoss is compatible: https://docs.newrelic.com/docs/agents/java-agent/getting-started/compatibility-requirements-java-agent#app-web-servers"))
				Expect(result.Status).To(Equal(tasks.Error))
			})
		})

		Context("When JBOSS_HOME is set and jbossAsReadme returns correctly", func() {
			BeforeEach(func() {

				upstream = map[string]tasks.Result{
					"Base/Env/CollectEnvVars": {
						Status:  tasks.Info,
						Payload: map[string]string{"JBOSS_HOME": "/foo/bar"}},
					"Java/Env/Process": {
						Status: tasks.Success,
					},
				}

				p.findFiles = func([]string, []string) []string {
					return []string{"README.txt"}
				}
				p.returnSubstringInFile = func(string, string) ([]string, error) {
					return []string{"JBoss Application Server 7.1.2"}, nil
				}
			})

			It("Should return Success status", func() {
				Expect(result.Status).To(Equal(tasks.Success))
			})

			It("Should indicate success in the result Summary", func() {
				Expect(result.Summary).To(ContainSubstring("JBoss version supported"))
			})

			It("Should include the version in the result Summary", func() {
				Expect(result.Summary).To(ContainSubstring("7.1.2"))
			})

		})
	})

	Describe("getHomeDirFromCmdline", func() {
		var (
			cmdLine string
			homeDir string
		)
		JustBeforeEach(func() {
			homeDir = p.getHomeDirFromCmdline(cmdLine)
		})
		Context("when linux cmdLine found", func() {
			BeforeEach(func() {
				cmdLine = "jboss.home.dir=/foo/bar"
			})
			It("Should return filtered homeDir for linux patterns", func() {
				Expect(homeDir).To(Equal("/foo/bar"))
			})

		})
		Context("when windows cmdLine", func() {
			BeforeEach(func() {
				cmdLine = `jboss.home.dir="C:\app\jboss-as-7.1.1.Final"`
			})
			It("Should return filtered homeDir for windows patterns", func() {
				Expect(homeDir).To(Equal(`"C:\app\jboss-as-7.1.1.Final"`))
			})
		})

		Context("when cmdLine doesn't contain jboss.home.dir", func() {
			BeforeEach(func() {
				cmdLine = ""
			})
			It("Should return empty string", func() {
				Expect(homeDir).To(Equal(""))
			})
		})

	})

	Describe("getAndParseJBossAsReadMe", func() {
		var (
			homepath      string
			versionString string
			err           error
		)
		JustBeforeEach(func() {
			versionString, err = p.getAndParseJBossAsReadMe(homepath)
		})
		Context("versionInHomePath length greater than zero", func() {
			BeforeEach(func() {
				homepath = "1.3.4"

			})
			It("Should return result from homepath directly", func() {
				Expect(versionString).To(Equal("1.3.4"))
				Expect(err).To(BeNil())
			})
		})
		Context("When readmes length is less than 0", func() {
			BeforeEach(func() {
				homepath = ""
				p.findFiles = func([]string, []string) []string {
					return []string{}
				}

			})
			It("Should return error result unable to find JBOSS readme", func() {
				Expect(versionString).To(Equal(""))
				Expect(err.Error()).To(Equal("error finding JBoss version"))
			})
		})
		Context("When versionStringRaw is less than 1", func() {
			BeforeEach(func() {
				homepath = ""
				p.findFiles = func([]string, []string) []string {
					return []string{"foo"}
				}
				p.returnSubstringInFile = func(string, string) ([]string, error) {
					return []string{}, nil
				}

			})
			It("Should return an error", func() {
				Expect(versionString).To(Equal(""))
				Expect(err.Error()).To(Equal("error finding version string"))
			})
		})
		Context("when versionStringRaw returned error", func() {
			BeforeEach(func() {
				homepath = ""
				p.findFiles = func([]string, []string) []string {
					return []string{"foo"}
				}
				p.returnSubstringInFile = func(string, string) ([]string, error) {
					return []string{}, errors.New("blargs")
				}
			})
			It("Should return an error", func() {
				Expect(versionString).To(Equal(""))
				Expect(err.Error()).To(Equal("error finding version string"))
			})
		})
		Context("versionString didn't contain at least 2 periods", func() {
			BeforeEach(func() {
				homepath = ""
				p.findFiles = func([]string, []string) []string {
					return []string{"foo"}
				}
				p.returnSubstringInFile = func(string, string) ([]string, error) {
					return []string{"", "1.2"}, nil
				}
			})
			It("Should return an error", func() {
				Expect(versionString).To(Equal(""))
				Expect(err.Error()).To(Equal("error finding version string"))
			})
		})
		Context("run checkJBossAsVersion function to get version", func() {
			BeforeEach(func() {
				homepath = ""
				p.findFiles = func([]string, []string) []string {
					return []string{"foo"}
				}
				p.returnSubstringInFile = func(string, string) ([]string, error) {
					return []string{"1.2.3"}, nil
				}
			})
			It("Should return mock summary", func() {
				Expect(versionString).To(Equal("1.2.3"))
				Expect(err).To(BeNil())
			})
		})
	})
	Describe("checkJBossAsVersion", func() {
		var (
			versionString string
			summary       string
			status        tasks.Status
		)
		JustBeforeEach(func() {
			summary, status = p.checkJBossAsVersion(versionString)
		})
		Context("when major version is not an int", func() {
			BeforeEach(func() {
				versionString = "foo.2.4"
			})
			It("Should return error result", func() {
				Expect(status).To(Equal(tasks.Error))
			})
		})
		Context("when minor version is not an int", func() {
			BeforeEach(func() {
				versionString = "2.foo.4"
			})
			It("Should", func() {
				Expect(status).To(Equal(tasks.Error))
			})
		})
		Context("when revision version is not an int", func() {
			BeforeEach(func() {
				versionString = "2.2.foo"
			})
			It("Should", func() {
				Expect(status).To(Equal(tasks.Error))
			})
		})
		Context("when major version is outside supported range", func() {
			BeforeEach(func() {
				versionString = "3.0.0"
			})
			It("Should return unsupported JBOSS version", func() {
				Expect(status).To(Equal(tasks.Failure))
				Expect(summary).To(Equal("Unsupported version of JBoss AS detected. Supported versions are 4.0.5 to AS 7.x. Detected version is 3.0.0"))
			})
		})
		Context("when jbossAS version is less than 4.0.5", func() {
			BeforeEach(func() {
				versionString = "4.0.4"
			})
			It("Should return unsupported result", func() {
				Expect(status).To(Equal(tasks.Failure))
				Expect(summary).To(Equal("Unsupported version of JBoss AS detected. Supported versions are 4.0.5 to AS 7.x. Detected version is 4.0.4"))
			})
		})
		Context("when supported jboss version found", func() {
			BeforeEach(func() {
				versionString = "5.2.4"
			})
			It("Should return successful result", func() {
				Expect(status).To(Equal(tasks.Success))
				Expect(summary).To(Equal("JBoss version supported. Version is 5.2.4"))
			})
		})
	})

})
