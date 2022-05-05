package daemon

// This is an example task test file referenced in /docs/unit-testing.md

import (
	"errors"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/shirou/gopsutil/process"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

func TestPHPDaemonRunning(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "PHP/Daemon/Running test suite")
}

// type processFinderFunc func(string) ([]process.Process, error)
// type fileExistsCheckerFunc func(string) bool

func mockProcessFinderZeroDaemon(name string) ([]process.Process, error) {
	return []process.Process{}, nil

}

func mockProcessFinderOneDaemon(name string) ([]process.Process, error) {
	return []process.Process{
		process.Process{Pid: 1},
	}, nil
}

func mockProcessFinderTwoDaemon(name string) ([]process.Process, error) {
	return []process.Process{
		process.Process{Pid: 1},
		process.Process{Pid: 2},
	}, nil
}

func mockProcessFinderThreeDaemon(name string) ([]process.Process, error) {
	return []process.Process{
		process.Process{Pid: 1},
		process.Process{Pid: 2},
		process.Process{Pid: 3},
	}, nil
}

func mockProcessFinderError(name string) ([]process.Process, error) {
	return []process.Process{}, errors.New("couldn't find daemon")
}

func mockFileExistsCheckerTrue(filename string) bool {
	return true
}

func mockFileExistsCheckerFalse(filename string) bool {
	return false
}

var _ = Describe("PHP/Daemon/Running", func() {
	var p PHPDaemonRunning

	Describe("Execute()", func() {
		var (
			result   tasks.Result
			options  tasks.Options
			upstream map[string]tasks.Result
		)

		Context("upstream dependency task succeeded", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"PHP/Config/Agent": tasks.Result{
						Status: tasks.Success,
					},
				}
			})

			Context("error finding daemon process", func() {
				JustBeforeEach(func() {
					p = PHPDaemonRunning{
						processFinder: mockProcessFinderError,
					}
					result = p.Execute(options, upstream)
				})

				It("should return an expected Error result status", func() {
					Expect(result.Status).To(Equal(tasks.Error))
				})

				It("should return an expected Error result summary including error", func() {
					Expect(result.Summary).To(Equal("Error detecting newrelic-daemon process: couldn't find daemon"))
				})
			})

			Context("Zero daemon processes found, agent start mode", func() {
				JustBeforeEach(func() {
					p = PHPDaemonRunning{
						processFinder:     mockProcessFinderZeroDaemon,
						fileExistsChecker: mockFileExistsCheckerFalse,
					}
					result = p.Execute(options, upstream)
				})

				It("should return an expected Failure result status", func() {
					Expect(result.Status).To(Equal(tasks.Failure))
				})

				It("should return a result summary with the number of daemon process", func() {
					Expect(result.Summary).To(Equal("There is incorrect number of newrelic-daemon processes running - (0). Please restart your web server to start up the daemons."))
				})
			})

			Context("Zero daemon processes found, manual start mode", func() {
				JustBeforeEach(func() {
					p = PHPDaemonRunning{
						processFinder:     mockProcessFinderZeroDaemon,
						fileExistsChecker: mockFileExistsCheckerTrue,
					}
					result = p.Execute(options, upstream)
				})

				It("should return an expected Failure result status", func() {
					Expect(result.Status).To(Equal(tasks.Failure))
				})

				It("should return a result summary with the number of daemon process", func() {
					Expect(result.Summary).To(Equal("There is incorrect number of newrelic-daemon processes running - (0). Please make sure the daemons were started as outlined in the following document:"))
				})

				It("should include a docs link outlining the different startup modes", func() {
					Expect(result.URL).To(Equal("https://docs.newrelic.com/docs/agents/php-agent/advanced-installation/starting-php-daemon-advanced"))
				})

			})

			Context("One daemon process found, manual start mode", func() {
				JustBeforeEach(func() {
					p = PHPDaemonRunning{
						processFinder:     mockProcessFinderOneDaemon,
						fileExistsChecker: mockFileExistsCheckerTrue,
					}
					result = p.Execute(options, upstream)
				})

				It("should return an expected Failure result status", func() {
					Expect(result.Status).To(Equal(tasks.Failure))
				})

				It("should return an expected Failure result summary", func() {
					Expect(result.Summary).To(Equal("There is incorrect number of newrelic-daemon processes running - (1). Please make sure the daemons were started as outlined in the following document:"))
				})

				It("should include a docs link outlining the different startup modes", func() {
					Expect(result.URL).To(Equal("https://docs.newrelic.com/docs/agents/php-agent/advanced-installation/starting-php-daemon-advanced"))
				})
			})

			Context("One daemon process found, agent start mode", func() {
				JustBeforeEach(func() {
					p = PHPDaemonRunning{
						processFinder:     mockProcessFinderOneDaemon,
						fileExistsChecker: mockFileExistsCheckerFalse,
					}
					result = p.Execute(options, upstream)
				})

				It("should return an expected Failure result status", func() {
					Expect(result.Status).To(Equal(tasks.Failure))
				})

				It("should return a result summary with the number of daemon process", func() {
					Expect(result.Summary).To(Equal("There is incorrect number of newrelic-daemon processes running - (1). Please restart your web server to start up the daemons."))
				})
			})

			Context("Two daemon processes found, agent start mode", func() {
				JustBeforeEach(func() {
					p = PHPDaemonRunning{
						processFinder:     mockProcessFinderTwoDaemon,
						fileExistsChecker: mockFileExistsCheckerFalse,
					}
					result = p.Execute(options, upstream)
				})

				It("should return an expected Success result status", func() {
					Expect(result.Status).To(Equal(tasks.Success))
				})

				It("should return a result summary with the number of daemon processes", func() {
					Expect(result.Summary).To(MatchRegexp("[Tt]wo"))
				})

				It("should return a result summary indicating agent start mode", func() {
					Expect(result.Summary).To(ContainSubstring("agent mode"))
				})

				It("Should return the expected payload", func() {
					expectedPayload := PHPDaemonInfo{
						ProcessCount: 2,
						Mode:         "agent",
					}
					Expect(result.Payload).To(Equal(expectedPayload))
				})
			})

			Context("Two daemon processes found, manual start mode", func() {
				JustBeforeEach(func() {
					p = PHPDaemonRunning{
						processFinder:     mockProcessFinderTwoDaemon,
						fileExistsChecker: mockFileExistsCheckerTrue,
					}
					result = p.Execute(options, upstream)
				})

				It("should return an expected Success result status", func() {
					Expect(result.Status).To(Equal(tasks.Success))
				})

				It("should return a result summary with the number of daemon processes", func() {
					Expect(result.Summary).To(MatchRegexp("[Tt]wo"))
				})

				It("should return a result summary indicating manual start mode", func() {
					Expect(result.Summary).To(ContainSubstring("manual mode"))
				})
				It("Should return the expected payload", func() {
					expectedPayload := PHPDaemonInfo{
						ProcessCount: 2,
						Mode:         "manual",
					}
					Expect(result.Payload).To(Equal(expectedPayload))
				})
			})

			Context("Three daemon process found, manual start mode", func() {
				JustBeforeEach(func() {
					p = PHPDaemonRunning{
						processFinder:     mockProcessFinderThreeDaemon,
						fileExistsChecker: mockFileExistsCheckerTrue,
					}
					result = p.Execute(options, upstream)
				})

				It("should return an expected Failure result status", func() {
					Expect(result.Status).To(Equal(tasks.Failure))
				})

				It("should return an expected Failure result summary", func() {
					Expect(result.Summary).To(Equal("There is incorrect number of newrelic-daemon processes running - (3). Please make sure the daemons were started as outlined in the following document:"))
				})

				It("should include a docs link outlining the different startup modes", func() {
					Expect(result.URL).To(Equal("https://docs.newrelic.com/docs/agents/php-agent/advanced-installation/starting-php-daemon-advanced"))
				})
			})

			Context("Three daemon process found, agent start mode", func() {
				JustBeforeEach(func() {
					p = PHPDaemonRunning{
						processFinder:     mockProcessFinderThreeDaemon,
						fileExistsChecker: mockFileExistsCheckerFalse,
					}
					result = p.Execute(options, upstream)
				})

				It("should return an expected Failure result status", func() {
					Expect(result.Status).To(Equal(tasks.Failure))
				})

				It("should return a result summary with the number of daemon process", func() {
					Expect(result.Summary).To(Equal("There is incorrect number of newrelic-daemon processes running - (3). Please restart your web server to start up the daemons."))
				})
				It("Should return the expected payload", func() {
					expectedPayload := PHPDaemonInfo{
						ProcessCount: 3,
						Mode:         "agent",
					}
					Expect(result.Payload).To(Equal(expectedPayload))
				})
			})
		})

		Context("Upstream dependency task failure", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"PHP/Config/Agent": tasks.Result{
						Status: tasks.Failure,
					},
				}
			})

			JustBeforeEach(func() {
				p = PHPDaemonRunning{}
				result = p.Execute(options, upstream)
			})

			It("should return an expected None result status", func() {
				Expect(result.Status).To(Equal(tasks.None))
			})

			It("should return a result summary indicating that it did not run", func() {
				Expect(result.Summary).To(Equal("PHP Agent was not detected on this host. Skipping daemon detection."))
			})
		})

		Context("Upstream dependency task warning", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"PHP/Config/Agent": tasks.Result{
						Status: tasks.Warning,
					},
				}
			})

			JustBeforeEach(func() {
				p = PHPDaemonRunning{}
				result = p.Execute(options, upstream)
			})

			It("should return an expected None result status", func() {
				Expect(result.Status).To(Equal(tasks.None))
			})

			It("should return a result summary indicating that it did not run", func() {
				Expect(result.Summary).To(Equal("PHP Agent was not detected on this host. Skipping daemon detection."))
			})
		})

		Context("Upstream dependency task error", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"PHP/Config/Agent": tasks.Result{
						Status: tasks.Error,
					},
				}
			})

			JustBeforeEach(func() {
				p = PHPDaemonRunning{}
				result = p.Execute(options, upstream)
			})

			It("should return an expected None result status", func() {
				Expect(result.Status).To(Equal(tasks.None))
			})

			It("should return a result summary indicating that it did not run", func() {
				Expect(result.Summary).To(Equal("PHP Agent was not detected on this host. Skipping daemon detection."))
			})
		})

		Context("Upstream dependency task none", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"PHP/Config/Agent": tasks.Result{
						Status: tasks.None,
					},
				}
			})

			JustBeforeEach(func() {
				p = PHPDaemonRunning{}
				result = p.Execute(options, upstream)
			})

			It("should return an expected None result status", func() {
				Expect(result.Status).To(Equal(tasks.None))
			})

			It("should return a result summary indicating that it did not run", func() {
				Expect(result.Summary).To(Equal("PHP Agent was not detected on this host. Skipping daemon detection."))
			})
		})

	})
})
